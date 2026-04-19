package storage

type ColumnName string

type UpdateColumn struct {
	ColumnName ColumnName
	Value      any
}

const (
	DailyRecordTotalSleepHour ColumnName = "total_sleep_hour"
	DailyRecordTotalWaterMl   ColumnName = "total_water_ml"
)

const (
	SendRequestStatus     ColumnName = "status"
	SendRequestFailReason ColumnName = "fail_reason"
)

const (
	UserMorningMsgSent ColumnName = "morning_msg_sent"
	UserEveningMsgSent ColumnName = "evening_msg_sent"
)
