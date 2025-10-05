package protocol

import "time"

const (
	CMD_KEY_REQ    = 1
	CMD_CONNECTION = 2
	CMD_REGISTER   = 3
	CMD_CHAT       = 4
	CMD_ROTATE_KEY = 5
	CMD_DELETE_ME  = 6

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

type messageContainer struct {
	UserName  []byte
	Signature []byte
	Message   []byte
}

func UnpackFrame(packedFrame []byte) (*messageContainer, error) {
	// TODO: Need to finish setting up the packedFrame parsing
	userName := packedFrame[:]
	message := packedFrame[:]
	signature := packedFrame[:]
	mc := &messageContainer{
		UserName:  userName,
		Message:   message,
		Signature: signature,
	}
	return mc, nil
}
