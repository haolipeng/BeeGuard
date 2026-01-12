package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	_ "embed"

	businessplugins "business_plugins/lib"

	_ "github.com/GehirnInc/crypt/sha256_crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
	"github.com/go-viper/mapstructure/v2"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/engine"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/utils"
	"go.uber.org/zap"
)

type UserHandler struct{}

func (*UserHandler) Name() string {
	return "user"
}

func (*UserHandler) DataType() int {
	return 5052
}

type utmp struct {
	Typ int16
	// alignment
	_    [2]byte
	Pid  int32
	Line [32]byte
	Id   [4]byte
	User [32]byte
	Host [256]byte
	Exit struct {
		Termination int16
		Exit        int16
	}
	Session int32
	Time    struct {
		Sec  int32
		Usec int32
	}
	Addr [16]byte
	// Reserved member
	Unused [20]byte
}

func maskPassword(pwd string) string {
	mpwd := []byte(pwd)
	switch len(pwd) {
	case 0:
		return ""
	case 1:
		return "*"
	case 2, 3:
		mpwd[1] = '*'
		return string(mpwd)
	case 4:
		mpwd[1] = '*'
		mpwd[2] = '*'
		return string(mpwd)
	default:
		for i := 2; i < len(mpwd)-2; i++ {
			mpwd[i] = '*'
		}
		return string(mpwd)
	}
}

type User struct {
	Username            string `mapstructure:"username"`
	Password            string `mapstructure:"password"`
	Uid                 string `mapstructure:"uid"`
	Gid                 string `mapstructure:"gid"`
	Groupname           string `mapstructure:"groupname"`
	Info                string `mapstructure:"info"`
	Home                string `mapstructure:"home"`
	Shell               string `mapstructure:"shell"`
	LastLoginTime       string `mapstructure:"last_login_time"`
	LastLoginIP         string `mapstructure:"last_login_ip"`
	WeakPassword        string `mapstructure:"weak_password"`
	WeakPasswordContent string `mapstructure:"weak_password_content"`
	Sudoers             string `mapstructure:"sudoers"`
}

func (h *UserHandler) Handle(c *businessplugins.Client, cache *engine.Cache, seq string) {
	//1.获取用户的基本信息(用户名、密码、登录shell等)
	f, err := os.Open("/etc/passwd")
	if err != nil {
		zap.S().Error(err)
	}
	m := map[string]*User{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		fields := strings.Split(s.Text(), ":")
		if len(fields) == 0 {
			continue
		}
		padding := len(fields)
		for i := 0; i < 7-padding; i++ {
			fields = append(fields, "")
		}
		u := &User{
			Username: fields[0],
			Password: fields[1],
			Uid:      fields[2],
			Gid:      fields[3],
			Info:     fields[4],
			Home:     fields[5],
			Shell:    fields[6],
		}
		u.Groupname, _ = utils.GetGroupname(fields[3])
		m[fields[0]] = u
	}
	f.Close()

	//2.获取用户的登录时间和登录ip
	f, err = os.Open("/var/log/wtmp")
	if err == nil {
		for {
			l := &utmp{}
			if er := binary.Read(f, binary.LittleEndian, l); er == nil {
				username := bytes.TrimRight(l.User[:], "\x00")
				ip := bytes.TrimRight(l.Addr[:], "\x00")
				//判断用户在映射表中是否存在
				if u, ok := m[string(username)]; ok {
					u.LastLoginIP = net.IP(ip).String()
					u.LastLoginTime = strconv.FormatInt(int64(l.Time.Sec), 10)
				}
			} else {
				break
			}
		}
		f.Close()
	}

	for _, u := range m {
		cmd := exec.Command("sudo", "-l", "-U", u.Username)
		output, err := cmd.CombinedOutput()
		if err == nil {
			if i := bytes.Index(output, []byte("may run the following commands")); i > 0 {
				output = output[i:]
				//定位到冒号，提取冒号后的权限内容，赋值为
				if i := bytes.IndexByte(output, ':'); i > 0 && len(output) > i+1 {
					output = output[i+1:]
					u.Sudoers = string(bytes.TrimSpace(output))
				}
			}
		}
		rec := &businessplugins.Record{
			DataType:  int32(h.DataType()),
			Timestamp: time.Now().Unix(),
			Data: &businessplugins.Payload{
				Fields: make(map[string]string, 12),
			},
		}
		//将User结构体转为map结构
		mapstructure.Decode(u, &rec.Data.Fields)

		//添加包序列号
		rec.Data.Fields["package_seq"] = seq
		c.SendRecord(rec)
	}
}
