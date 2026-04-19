package msg

import (
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
)

type MsgType int

const (
	MsgTypeUnknown   MsgType = 0
	MsgTypeMorning   MsgType = 1
	MsgTypeEvening   MsgType = 2
	MsgTypeBreakfast MsgType = 3
	MsgTypeLunch     MsgType = 4
	MsgTypeDinner    MsgType = 5
)

func (m MsgType) GetColumn() string {
	switch m {
	case MsgTypeMorning:
		return "morning_msg_sent"
	case MsgTypeEvening:
		return "evening_msg_sent"
	case MsgTypeBreakfast:
		return "breakfast_msg_sent"
	case MsgTypeLunch:
		return "lunch_msg_sent"
	case MsgTypeDinner:
		return "dinner_msg_sent"
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
	case MsgTypeBreakfast:
		return models.SendRequestTypePushBreakfastMsg
	case MsgTypeLunch:
		return models.SendRequestTypePushLunchMsg
	case MsgTypeDinner:
		return models.SendRequestTypePushDinnerMsg
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
	case "breakfast":
		return MsgTypeBreakfast
	case "lunch":
		return MsgTypeLunch
	case "dinner":
		return MsgTypeDinner
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
	case MsgTypeBreakfast:
		return "breakfast"
	case MsgTypeLunch:
		return "lunch"
	case MsgTypeDinner:
		return "dinner"
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
	case MsgTypeBreakfast:
		return "早餐"
	case MsgTypeLunch:
		return "午餐"
	case MsgTypeDinner:
		return "晚餐"
	default:
		return ""
	}
}

func (m MsgType) GetMeal() models.Meal {
	switch m {
	case MsgTypeBreakfast:
		return models.MealBreakfast
	case MsgTypeLunch:
		return models.MealLunch
	case MsgTypeDinner:
		return models.MealDinner
	default:
		return models.MealUnknown
	}
}
