//go:generate protoc --gogofaster_out=:. bridge.proto
package businessplugins

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
	"sync"
	"time"
)

type Client struct {
	rx     io.ReadCloser
	tx     io.WriteCloser
	reader *bufio.Reader
	writer *bufio.Writer
	rmu    *sync.Mutex
	wmu    *sync.Mutex
}

// New 创建新的 Client，从标准输入/输出和文件描述符初始化
func New() (c *Client) {
	c = &Client{
		rx: os.Stdin,
		tx: os.Stdout,
		// MAX_SIZE = 1 MB
		reader: bufio.NewReaderSize(os.NewFile(3, "pipe"), 1024*1024),
		writer: bufio.NewWriterSize(os.NewFile(4, "pipe"), 512*1024),
		rmu:    &sync.Mutex{},
		wmu:    &sync.Mutex{},
	}
	go func() {
		ticker := time.NewTicker(time.Millisecond * 200)
		defer ticker.Stop()
		for {
			<-ticker.C
			if err := c.Flush(); err != nil {
				break
			}
		}
	}()
	return
}

func (c *Client) SendRecord(rec *Record) (err error) {
	c.wmu.Lock()
	defer c.wmu.Unlock()

	//写入记录大小
	size := rec.Size()
	err = binary.Write(c.writer, binary.LittleEndian, uint32(size))
	if err != nil {
		return err
	}

	//序列化记录
	var buf []byte
	buf, err = rec.Marshal()
	if err != nil {
		return err
	}

	//写入记录
	_, err = c.writer.Write(buf)
	if err != nil {
		return err
	}

	return
}

func (c *Client) ReceiveTask() (t *Task, err error) {
	c.rmu.Lock()
	defer c.rmu.Unlock()

	//读取任务大小
	var len uint32
	err = binary.Read(c.reader, binary.LittleEndian, &len)
	if err != nil {
		return
	}

	//读取任务详情
	var buf []byte
	buf, err = c.reader.Peek(int(len))
	if err != nil {
		return
	}

	// 丢弃任务详情
	_, err = c.reader.Discard(int(len))
	if err != nil {
		return
	}

	//反序列化任务
	t = &Task{}
	err = t.Unmarshal(buf)

	return
}

func (c *Client) Flush() (err error) {
	c.wmu.Lock()
	defer c.wmu.Unlock()

	//如果缓冲区有数据，则刷新缓冲区
	if c.writer.Buffered() != 0 {
		err = c.writer.Flush()
	}

	return
}

func (c *Client) Close() {
	c.writer.Flush()
	c.rx.Close()
	c.tx.Close()
}
