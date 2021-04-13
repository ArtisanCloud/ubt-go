package UBT

import (
	"os"
	"os/user"
)

func getSystemInfo() map[string]string {
	systemInfo := map[string]string{}

	// 获取主机名
	hostname, err := os.Hostname()
	if err == nil {
		systemInfo["hostname"] = hostname
	}

	// 获取用户名
	currentUser, err := user.Current()
	if err == nil {
		systemInfo["username"] = currentUser.Name
	}

	return systemInfo
}
