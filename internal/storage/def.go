package storage

type ColumnName string

type UpdateColumn struct {
	ColumnName ColumnName
	Value      any
}

const (
	DailyRecordTotalSleepHour ColumnName = "total_sleep_hour"
	DailyRecordTotalWaterMl   ColumnName = "total_water_ml"
	DailyRecordBreakfastMeals ColumnName = "breakfast_meals"
	DailyRecordLunchMeals     ColumnName = "lunch_meals"
	DailyRecordDinnerMeals    ColumnName = "dinner_meals"
	DailyRecordSnackMeals     ColumnName = "snack_meals"
)

const (
	SendRequestStatus     ColumnName = "status"
	SendRequestFailReason ColumnName = "fail_reason"
)

const (
	UserMorningMsgSent   ColumnName = "morning_msg_sent"
	UserEveningMsgSent   ColumnName = "evening_msg_sent"
	UserBreakfastMsgSent ColumnName = "breakfast_msg_sent"
	UserLunchMsgSent     ColumnName = "lunch_msg_sent"
	UserDinnerMsgSent    ColumnName = "dinner_msg_sent"
	UserTargetTimeframe  ColumnName = "target_timeframe"
	UserTargetWeight     ColumnName = "target_weight"
)
