package action

type PostbackActionType int

const (
	ActionTypeUnknown PostbackActionType = iota
	ActionTypeSetMeal
	ActionTypeCheckDescript
	ActionTypeSetDescription
	ActionTypeCheckBasicInfo
	ActionTypeDailyReport
	ActionTypeChangeTarget
	ActionTypeSetWater
	ActionTypeSetSleep
	ActionTypeSetWeight
	ActionTypeSetting
	// ActionTypeJoinUs
	// ActionTypeGetMonth
	// ActionTypeSetReportStart
	// ActionTypeSetReportEnd
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
	// case ActionTypeJoinUs:
	// 	return "join_us"
	// case ActionTypeGetMonth:
	// 	return "get_month"
	// case ActionTypeSetReportStart:
	// 	return "set_report_start"
	// case ActionTypeSetReportEnd:
	// 	return "set_report_end"
	case ActionTypeChangeTarget:
		return "change_target"
	case ActionTypeSetWater:
		return "set_water"
	case ActionTypeSetSleep:
		return "set_sleep"
	case ActionTypeSetWeight:
		return "set_weight"
	case ActionTypeSetting:
		return "setting"
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
	// case "join_us":
	// 	return ActionTypeJoinUs
	// case "get_month":
	// 	return ActionTypeGetMonth
	// case "set_report_start":
	// 	return ActionTypeSetReportStart
	// case "set_report_end":
	// 	return ActionTypeSetReportEnd
	case "change_target":
		return ActionTypeChangeTarget
	case "set_water":
		return ActionTypeSetWater
	case "set_sleep":
		return ActionTypeSetSleep
	case "set_weight":
		return ActionTypeSetWeight
	case "setting":
		return ActionTypeSetting
	default:
		return ActionTypeUnknown
	}
}

type TextMessageActionType int

const (
	TextMessageActionTypeUnknown TextMessageActionType = iota
	TextMessageActionTypeSetDescription
	TextMessageActionTypeSetTarget
	TextMessageActionTypeRecordWater
	TextMessageActionTypeRecordSleep
	TextMessageActionTypeRecordWeight
)

func (a TextMessageActionType) String() string {
	switch a {
	case TextMessageActionTypeSetDescription:
		return "set_description"
	case TextMessageActionTypeSetTarget:
		return "set_target"
	case TextMessageActionTypeRecordWater:
		return "record_water"
	case TextMessageActionTypeRecordSleep:
		return "record_sleep"
	case TextMessageActionTypeRecordWeight:
		return "record_weight"
	default:
		return "unknown"
	}
}

func ToTextMessageActionType(s string) TextMessageActionType {
	switch s {
	case "set_description":
		return TextMessageActionTypeSetDescription
	case "set_target":
		return TextMessageActionTypeSetTarget
	case "record_water":
		return TextMessageActionTypeRecordWater
	case "record_sleep":
		return TextMessageActionTypeRecordSleep
	case "record_weight":
		return TextMessageActionTypeRecordWeight
	default:
		return TextMessageActionTypeUnknown
	}
}
