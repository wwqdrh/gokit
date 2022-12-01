package ssh

import (
	"fmt"
	"io"
	"os"
)

const (
	// SshBitSize ssh bit size
	SshBitSize      = 2048
	StandardSshPort = 22
	// SshAuthKey auth key name
	SshAuthKey = "authorized"
	// SshAuthPrivateKey ssh private key
	SshAuthPrivateKey = "privateKey"
	// PostfixRsaKey postfix of local private key name
	PostfixRsaKey = ".key"
	Eol           = "\n"
)

var BackgroundLogger = io.Discard

var (
	UserHome      = os.Getenv("HOME")
	NettoolHome   = fmt.Sprintf("%s/.nettool", UserHome)
	NettoolKeyDir = fmt.Sprintf("%s/key", NettoolHome)
)
