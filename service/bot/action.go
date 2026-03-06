package bot

type ActionType int

const (
	ActionTypeUnknown ActionType = iota
	ActionTypeSetMeal
	ActionTypeCheckDescript
	ActionTypeSetDescription
	ActionTypeCheckBasicInfo
	ActionTypeDailyReport
	ActionTypeJoinUs
	ActionTypeGetMonth
	ActionTypeSetReportStart
	ActionTypeSetReportEnd
)

func (a ActionType) String() string {
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
	default:
		return "unknown"
	}
}

func ToActionType(s string) ActionType {
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
	default:
		return ActionTypeUnknown
	}
}
