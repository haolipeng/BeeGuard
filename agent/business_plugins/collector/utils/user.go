package utils

import (
	"os/user"
)

// GetUsername 根据 UID 获取用户名（简化版本，不使用缓存）
func GetUsername(uid string) (ret string, err error) {
	var u *user.User
	u, err = user.LookupId(uid)
	if err != nil {
		return
	}
	ret = u.Username
	return
}

// GetGroupname 根据 GID 获取组名（简化版本，不使用缓存）
func GetGroupname(gid string) (ret string, err error) {
	var g *user.Group
	g, err = user.LookupGroupId(gid)
	if err != nil {
		return
	}
	ret = g.Name
	return
}
