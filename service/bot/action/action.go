package action

type PostbackActionType int

const (
	ActionTypeUnknown PostbackActionType = iota
	ActionTypeSetMeal
	ActionTypeCheckDescript
	ActionTypeSetDescription
	ActionTypeCheckBasicInfo
	ActionTypeDailyReport
	ActionTypeJoinUs
	ActionTypeGetMonth
	ActionTypeSetReportStart
	ActionTypeSetReportEnd
	ActionTypeChangeTargetWeight
	ActionTypeSetWater
	ActionTypeSetSleep
)

func (a PostbackActionType) String() string {
	switch a {
	case ActionTypeSetMeal:
		return "set_meal"
	case ActionTypeCheckDescript:
		return "check_descript"
	case ActionTypeSetDescription:
		return "set_description"
	case ActionTypeCheckBasicInfo:
		return "check_basic_info"
	case ActionTypeDailyReport:
		return "daily_report"
	case ActionTypeJoinUs:
		return "join_us"
	case ActionTypeGetMonth:
		return "get_month"
	case ActionTypeSetReportStart:
		return "set_report_start"
	case ActionTypeSetReportEnd:
		return "set_report_end"
	case ActionTypeChangeTargetWeight:
		return "change_target_weight"
	case ActionTypeSetWater:
		return "set_water"
	case ActionTypeSetSleep:
		return "set_sleep"
	default:
		return "unknown"
	}
}

func ToPostbackActionType(s string) PostbackActionType {
	switch s {
	case "set_meal":
		return ActionTypeSetMeal
	case "check_descript":
		return ActionTypeCheckDescript
	case "set_description":
		return ActionTypeSetDescription
	case "check_basic_info":
		return ActionTypeCheckBasicInfo
	case "daily_report":
		return ActionTypeDailyReport
	case "join_us":
		return ActionTypeJoinUs
	case "get_month":
		return ActionTypeGetMonth
	case "set_report_start":
		return ActionTypeSetReportStart
	case "set_report_end":
		return ActionTypeSetReportEnd
	case "change_target_weight":
		return ActionTypeChangeTargetWeight
	case "set_water":
		return ActionTypeSetWater
	case "set_sleep":
		return ActionTypeSetSleep
	default:
		return ActionTypeUnknown
	}
}

type TextMessageActionType int

const (
	TextMessageActionTypeUnknown TextMessageActionType = iota
	TextMessageActionTypeSetDescription
	TextMessageActionTypeSetTargetWeight
	TextMessageActionTypeRecordWater
	TextMessageActionTypeRecordSleep
)

func (a TextMessageActionType) String() string {
	switch a {
	case TextMessageActionTypeSetDescription:
		return "set_description"
	case TextMessageActionTypeSetTargetWeight:
		return "set_target_weight"
	case TextMessageActionTypeRecordWater:
		return "record_water"
	case TextMessageActionTypeRecordSleep:
		return "record_sleep"
	default:
		return "unknown"
	}
}

func ToTextMessageActionType(s string) TextMessageActionType {
	switch s {
	case "set_description":
		return TextMessageActionTypeSetDescription
	case "set_target_weight":
		return TextMessageActionTypeSetTargetWeight
	case "record_water":
		return TextMessageActionTypeRecordWater
	case "record_sleep":
		return TextMessageActionTypeRecordSleep
	default:
		return TextMessageActionTypeUnknown
	}
}
