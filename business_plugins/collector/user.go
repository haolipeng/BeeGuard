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
	// 基本信息
	Username      string `mapstructure:"username"`
	Password      string `mapstructure:"password"`
	Uid           string `mapstructure:"uid"`
	Gid           string `mapstructure:"gid"`
	Groupname     string `mapstructure:"groupname"`
	Info          string `mapstructure:"info"`
	Home          string `mapstructure:"home"`
	Shell         string `mapstructure:"shell"`
	LastLoginTime string `mapstructure:"last_login_time"`
	LastLoginIP   string `mapstructure:"last_login_ip"`

	// 账号类型标识（布尔值，便于查询和过滤）
	IsRoot         string `mapstructure:"is_root"`          // "true" 或 "false"，UID == "0" 时为 "true"
	IsSudo         string `mapstructure:"is_sudo"`          // "true" 或 "false"，有 sudo 权限时为 "true"
	IsExpired      string `mapstructure:"is_expired"`       // "true" 或 "false"，密码已过期时为 "true"
	IsExpiringSoon string `mapstructure:"is_expiring_soon"` // "true" 或 "false"，密码即将过期时为 "true"

	// 密码过期详细信息（从 /etc/shadow 获取）
	PasswordLastChange string `mapstructure:"password_last_change"` // 密码最后修改日期（Unix 时间戳，秒）
	PasswordMaxDays    string `mapstructure:"password_max_days"`    // 密码最大使用天数（0=永不过期）
	PasswordWarnDays   string `mapstructure:"password_warn_days"`   // 密码过期前警告天数
	PasswordExpireDate string `mapstructure:"password_expire_date"` // 密码过期日期（Unix 时间戳，秒）
	PasswordRemainDays string `mapstructure:"password_remain_days"` // 密码剩余有效天数（负数表示已过期）

	// Sudo 权限信息
	Sudoers string `mapstructure:"sudoers"` // sudo 权限内容（有权限时才有值）

	// 弱密码检测（可选，如果需要的话）
	//WeakPassword        string `mapstructure:"weak_password"`
	//WeakPasswordContent string `mapstructure:"weak_password_content"`
}

// checkIsRoot 判断是否为 root 账号（UID == "0"）
func checkIsRoot(uid string) string {
	if uid == "0" {
		return "true"
	}
	return "false"
}

// checkIsSudo 判断是否有 sudo 权限（通过检查 Sudoers 字段）
func checkIsSudo(sudoers string) string {
	if sudoers != "" {
		return "true"
	}
	return "false"
}

// parseShadowFile 解析 /etc/shadow 文件，更新用户的密码过期信息
// shadow 文件格式：username:password:last_change:min_days:max_days:warn_days:inactive_days:expire_date:reserved
func parseShadowFile(m map[string]*User) {
	f, err := os.Open("/etc/shadow")
	if err != nil {
		// /etc/shadow 文件需要 root 权限才能读取，如果无法读取则跳过
		zap.S().Warnf("Failed to open /etc/shadow: %v (may need root privileges)", err)
		return
	}
	defer f.Close()

	// 获取当前时间（Unix 时间戳，秒）
	now := time.Now().Unix()
	// 1970年1月1日的时间戳（Unix epoch）
	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	// 当前日期距离1970年1月1日的天数
	todayDays := int64(now-epoch) / 86400

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过注释和空行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) < 3 {
			continue
		}

		username := fields[0]
		// 查找对应的用户
		u, ok := m[username]
		if !ok {
			continue
		}

		// 解析密码最后修改日期（字段2：从1970年1月1日算起的天数）
		lastChangeDays := int64(0)
		if fields[2] != "" {
			if days, err := strconv.ParseInt(fields[2], 10, 64); err == nil {
				lastChangeDays = days
				// 转换为 Unix 时间戳（秒）
				lastChangeTimestamp := epoch + days*86400
				u.PasswordLastChange = strconv.FormatInt(lastChangeTimestamp, 10)
			}
		}

		// 解析密码最大使用天数（字段4：0或99999表示永不过期）
		maxDays := int64(0)
		if len(fields) > 4 && fields[4] != "" {
			if days, err := strconv.ParseInt(fields[4], 10, 64); err == nil {
				maxDays = days
				u.PasswordMaxDays = strconv.FormatInt(days, 10)
			}
		}

		// 解析密码过期前警告天数（字段5）
		warnDays := int64(0)
		if len(fields) > 5 && fields[5] != "" {
			if days, err := strconv.ParseInt(fields[5], 10, 64); err == nil {
				warnDays = days
				u.PasswordWarnDays = strconv.FormatInt(days, 10)
			}
		}

		// 计算密码过期信息
		// 如果 maxDays == 0 或 maxDays >= 99999，表示密码永不过期
		if maxDays == 0 || maxDays >= 99999 {
			u.IsExpired = "false"
			u.PasswordExpireDate = "0"     // 0 表示永不过期
			u.PasswordRemainDays = "99999" // 表示永不过期
			continue
		}

		// 计算密码过期日期（天数）
		expireDays := lastChangeDays + maxDays
		// 计算密码过期日期（Unix 时间戳，秒）
		expireTimestamp := epoch + expireDays*86400
		u.PasswordExpireDate = strconv.FormatInt(expireTimestamp, 10)

		// 计算密码剩余有效天数
		remainDays := expireDays - todayDays
		u.PasswordRemainDays = strconv.FormatInt(remainDays, 10)

		// 判断密码是否已过期
		if remainDays < 0 {
			u.IsExpired = "true"
			u.IsExpiringSoon = "false" // 已过期不算即将过期
		} else {
			u.IsExpired = "false"
			// 判断密码是否即将过期（剩余天数 <= 警告天数）
			if warnDays > 0 && remainDays <= warnDays {
				u.IsExpiringSoon = "true"
			} else {
				u.IsExpiringSoon = "false"
			}
		}
	}

	if err := scanner.Err(); err != nil {
		zap.S().Warnf("Error reading /etc/shadow: %v", err)
	}
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
		// 判断是否为 root 账号
		u.IsRoot = checkIsRoot(fields[2])
		// 初始化其他字段的默认值
		u.IsSudo = "false"
		u.IsExpired = "false"
		u.IsExpiringSoon = "false"
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

	//3.检查 sudo 权限
	for _, u := range m {
		cmd := exec.Command("sudo", "-l", "-U", u.Username)
		output, err := cmd.CombinedOutput()
		if err == nil {
			if i := bytes.Index(output, []byte("may run the following commands")); i > 0 {
				output = output[i:]
				//定位到冒号，提取冒号后的权限内容
				if i := bytes.IndexByte(output, ':'); i > 0 && len(output) > i+1 {
					output = output[i+1:]
					u.Sudoers = string(bytes.TrimSpace(output))
				}
			}
		}
		// 判断是否有 sudo 权限
		u.IsSudo = checkIsSudo(u.Sudoers)
	}

	//4.检查密码过期信息（从 /etc/shadow 读取）
	parseShadowFile(m)

	//5.发送记录
	for _, u := range m {
		rec := &businessplugins.Record{
			DataType:  int32(h.DataType()),
			Timestamp: time.Now().Unix(),
			Data: &businessplugins.Payload{
				Fields: make(map[string]string, 20), // 增加字段数量以容纳新字段
			},
		}
		//将User结构体转为map结构
		mapstructure.Decode(u, &rec.Data.Fields)

		//添加包序列号
		rec.Data.Fields["package_seq"] = seq
		c.SendRecord(rec)
	}
}
