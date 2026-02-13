package plugin

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sync"
	"syscall"
	"time"

	"gitlab.myinterest.top/security/agent/agent"
	"gitlab.myinterest.top/security/agent/buffer"
	"gitlab.myinterest.top/security/agent/proto"
	"go.uber.org/zap"
)

// 插件关闭
func (p *Plugin) Shutdown() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.IsExited() {
		return
	}

	p.Info("plugin is running, will shutdown it")

	// 先发 SIGTERM 通知插件优雅退出
	syscall.Kill(-p.cmd.Process.Pid, syscall.SIGTERM)
	p.tx.Close()
	p.rx.Close()
	select {
	case <-time.After(time.Second * 10):
		p.Warn("because of plugin exit's timeout, will kill it")
		syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
		<-p.done
		p.Info("plugin has been killed")
	case <-p.done:
		p.Info("plugin has been shutdown gracefully")
	}
}

// 插件加载
func Load(ctx context.Context, config proto.Config) (plg *Plugin, err error) {
	loadedPlg, ok := m.Load(config.Name)
	if ok {
		loadedPlg := loadedPlg.(*Plugin)
		//已经加载相同版本的插件
		if loadedPlg.Config.Version == config.Version && loadedPlg.cmd.ProcessState == nil {
			err = ErrDuplicatePlugin
			return
		}
		//插件版本不同，关闭旧版本
		if loadedPlg.Config.Version != config.Version && loadedPlg.cmd.ProcessState == nil {
			loadedPlg.Infof("because of the different plugin's version,the previous version will be shutdown...")
			loadedPlg.Shutdown()
			loadedPlg.Infof("shutdown successfully")
		}
	}
	if config.Signature == "" {
		config.Signature = config.Sha256
	}
	logger := zap.S().With("plugin", config.Name, "pver", config.Version, "psign", config.Signature)
	logger.Info("plugin is loading...")

	workingDirectory := path.Join(agent.PluginsDirectory, config.Name)
	// for compatibility
	os.Remove(path.Join(workingDirectory, config.Name+".stderr"))
	os.Remove(path.Join(workingDirectory, config.Name+".stdout"))
	execPath := path.Join(workingDirectory, config.Name)

	// 注意：这里需要 utils.CheckSignature 和 utils.Download，暂时注释掉，后续需要实现 utils 模块
	// err = utils.CheckSignature(execPath, config.Signature)
	// if err != nil {
	// 	logger.Warn("check local plugin's signature failed: ", err)
	// 	logger.Info("downloading plugin from remote server...")
	// 	err = utils.Download(ctx, execPath, config)
	// 	if err != nil {
	// 		return
	// 	}
	// 	logger.Info("download done")
	// }

	// 检查插件文件是否存在
	if _, err = os.Stat(execPath); os.IsNotExist(err) {
		logger.Errorf("plugin executable not found: %s", execPath)
		return nil, err
	}

	// 创建进程间通信的管道
	cmd := exec.Command(execPath)
	var rx_r, rx_w, tx_r, tx_w *os.File

	// Agent接收插件数据的管道
	rx_r, rx_w, err = os.Pipe()
	if err != nil {
		return
	}

	// Agent发送任务给插件的管道
	tx_r, tx_w, err = os.Pipe()
	if err != nil {
		return
	}

	//配置进程属性
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.ExtraFiles = append(cmd.ExtraFiles, tx_r, rx_w)
	cmd.Dir = workingDirectory
	var errFile *os.File
	errFile, err = os.OpenFile(execPath+".stderr", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o0600)
	if err != nil {
		return
	}
	defer errFile.Close()
	cmd.Stderr = errFile
	if config.Detail != "" {
		//传递环境变量
		cmd.Env = append(cmd.Env, "DETAIL="+config.Detail)
	}
	// 传递日志目录环境变量
	logDir := path.Join(agent.LogDirectory, "plugins", config.Name)
	cmd.Env = append(cmd.Env, "LOG_DIR="+logDir)

	logger.Info("plugin's process will start")

	// 启动插件进程
	err = cmd.Start()
	tx_r.Close()
	rx_w.Close()
	if err != nil {
		return
	}
	// 判断插件是否使用标准protobuf协议（driver等eBPF插件）
	useStandardProtocol := false
	if config.Name == "ebpf_base_detector" {
		useStandardProtocol = true
		logger.Info("plugin will use standard protobuf protocol")
	}

	plg = &Plugin{
		Config:              config,
		mu:                  &sync.Mutex{},
		cmd:                 cmd,
		rx:                  rx_r,
		updateTime:          time.Now(),
		reader:              bufio.NewReaderSize(rx_r, 1024*128),
		tx:                  tx_w,
		done:                make(chan struct{}),
		taskCh:              make(chan proto.Task),
		wg:                  &sync.WaitGroup{},
		useStandardProtocol: useStandardProtocol,
		SugaredLogger:       logger,
	}
	plg.wg.Add(3)

	// 协程1: 等待插件进程退出
	go func() {
		defer plg.wg.Done()
		defer plg.Info("gorountine of waiting plugin's process will exit")

		//等待进程退出
		err = cmd.Wait()
		rx_r.Close()
		tx_w.Close()
		if err != nil {
			plg.Errorf("plugin has exited with error:%v,code:%d", err, cmd.ProcessState.ExitCode())
		} else {
			plg.Infof("plugin has exited with code %d", cmd.ProcessState.ExitCode())
		}
		//关闭插件done通道
		close(plg.done)
	}()

	// 协程2: 接收插件数据
	go func() {
		defer plg.wg.Done()
		defer plg.Info("gorountine of receiving plugin's data will exit")

		for {
			var rec *proto.EncodedRecord
			var err error

			// 根据插件类型选择接收方法
			if plg.useStandardProtocol {
				rec, err = plg.ReceiveStandardRecord()
			} else {
				rec, err = plg.ReceiveData()
			}

			if err != nil {
				if errors.Is(err, bufio.ErrBufferFull) {
					plg.Warn("when receiving data, buffer is full, skip this record")
					continue
				} else if !(errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrClosed)) {
					plg.Error("when receiving data, an error occurred: ", err)
				} else {
					plg.Info("when receiving data, pipe closed normally")
					break
				}
			}
			buffer.WriteEncodedRecord(rec)
		}
	}()

	// 协程3: 发送任务给插件
	go func() {
		defer plg.wg.Done()
		defer plg.Info("gorountine of sending task to plugin will exit")

		for {
			select {
			case <-plg.done:
				plg.Info("plugin's done channel has been closed, will exit")
				return
			case task := <-plg.taskCh: //任务通道有任务
				s := task.Size()
				var dst = make([]byte, 4+s)                 //分配缓冲区，前四个字节保存长度
				_, err = task.MarshalToSizedBuffer(dst[4:]) //将任务序列化到缓冲区
				if err != nil {
					plg.Errorf("when marshaling a task, an error occurred: %v, ignored this task: %+v", err, task)
					continue
				}
				binary.LittleEndian.PutUint32(dst[:4], uint32(s))
				var n int
				n, err = plg.tx.Write(dst)
				if err != nil {
					if !errors.Is(err, os.ErrClosed) {
						plg.Error("when sending task, an error occurred: ", err)
					}
					return
				}
				fmt.Println("send task", n)
				//atomic.AddUint64(&plg.rxCnt, 1)
				//atomic.AddUint64(&plg.rxBytes, uint64(n))
			}
		}
	}()
	m.Store(config.Name, plg)
	return
}
