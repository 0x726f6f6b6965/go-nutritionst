package msg

import (
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
)

type MsgType int

const (
	MsgTypeUnknown MsgType = 0
	MsgTypeMorning MsgType = 1
	MsgTypeEvening MsgType = 2
)

func (m MsgType) GetColumn() string {
	switch m {
	case MsgTypeMorning:
		return "morning_msg_sent"
	case MsgTypeEvening:
		return "evening_msg_sent"
	default:
		return ""
	}
}

func (m MsgType) GetRequestType() models.SendRequestType {
	switch m {
	case MsgTypeMorning:
		return models.SendRequestTypePushMorningMsg
	case MsgTypeEvening:
		return models.SendRequestTypePushEveningMsg
	default:
		return models.SendRequestTypeUnknown
	}
}

func GetMsgType(s string) MsgType {
	switch s {
	case "morning":
		return MsgTypeMorning
	case "evening":
		return MsgTypeEvening
	default:
		return MsgTypeUnknown
	}
}

func (m MsgType) String() string {
	switch m {
	case MsgTypeMorning:
		return "morning"
	case MsgTypeEvening:
		return "evening"
	default:
		return "unknown"
	}
}

func (m MsgType) ChineseString() string {
	switch m {
	case MsgTypeMorning:
		return "早上"
	case MsgTypeEvening:
		return "晚間"
	default:
		return ""
	}
}
