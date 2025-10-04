package protocol

import "time"

const (
	CMD_KEY_REQ    = 0
	CMD_CONNECTION = 1
	CMD_REGISTER   = 2
	CMD_CHAT       = 3
	CMD_ROTATE_KEY = 4
	CMD_DELETE_ME  = 5

	KeepAliveTimer = 30 * time.Second
	DefaultPort    = 8080

	ErrGenericCrash = 1
	Protocol        = "tcp"

	//Frame Byte Structure
	Command      = 1
	UserName     = 32
	Signature    = 64
	Message      = 911
	PayloadTotal = 1024
)
