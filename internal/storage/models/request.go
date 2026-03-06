package models

import "time"

type SendRequest struct {
	ID          int64
	RequestID   string
	RequestType SendRequestType
	Status      SendRequestStatus
	Error       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SendRequestType int

type SendRequestStatus int

const (
	SendRequestTypeUnknown SendRequestType = 0
	SendRequestTypeMeal    SendRequestType = 1
	SendRequestTypePlan    SendRequestType = 2
)

const (
	SendRequestStatusUnknown SendRequestStatus = 0
	SendRequestStatusPending SendRequestStatus = 1
	SendRequestStatusSuccess SendRequestStatus = 2
	SendRequestStatusFailed  SendRequestStatus = 3
)
