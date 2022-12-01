package ostool

import (
	"os"
	"os/user"
)

func IsRunAsAdmin() bool {
	return os.Geteuid() == 0
}

func GetAdminUserName() string {
	return "root"
}

// GetLocalUserName get current username
func GetLocalUserName() string {
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" {
		return sudoUser
	}
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.Username
}
