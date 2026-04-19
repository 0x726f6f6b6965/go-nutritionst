package models

import "time"

type SendRequest struct {
	ID          int64             `json:"id" db:"id"`
	RequestID   string            `json:"request_id" db:"request_id"`
	LineID      string            `json:"line_id" db:"line_id"`
	RequestType SendRequestType   `json:"request_type" db:"request_type"`
	Status      SendRequestStatus `json:"status" db:"status"`
	Data        string            `json:"data" db:"data"`
	Error       string            `json:"error" db:"fail_reason"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

type SendRequestType int

type SendRequestStatus int

const (
	SendRequestTypeUnknown          SendRequestType = 0
	SendRequestTypeMeal             SendRequestType = 1
	SendRequestTypeDaily            SendRequestType = 2
	SendRequestTypeBroadcast        SendRequestType = 3
	SendRequestTypeBasicInfo        SendRequestType = 4
	SendRequestTypePushMorningMsg   SendRequestType = 5
	SendRequestTypePushEveningMsg   SendRequestType = 6
	SendRequestTypePushBreakfastMsg SendRequestType = 7
	SendRequestTypePushLunchMsg     SendRequestType = 8
	SendRequestTypePushDinnerMsg    SendRequestType = 9
)

const (
	SendRequestStatusUnknown SendRequestStatus = 0
	SendRequestStatusPending SendRequestStatus = 1
	SendRequestStatusSuccess SendRequestStatus = 2
	SendRequestStatusFailed  SendRequestStatus = 3
)
