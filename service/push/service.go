package push

import (
	"context"
	"fmt"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/query"
	"github.com/0x726f6f6b6965/go-nutritionst/service/push/msg"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

const (
	Limit = 1000
)

type Service struct {
	lineClient *messaging_api.MessagingApiAPI
	store      *storage.Postgres
}

type PushMsgRequest struct {
	Msg string
	Typ msg.MsgType
}

func NewService(lineClient *messaging_api.MessagingApiAPI, db *storage.Postgres) *Service {
	return &Service{
		lineClient: lineClient,
		store:      db,
	}
}

func (s *Service) PushMsg(ctx context.Context, req *PushMsgRequest) error {
	column := req.Typ.GetColumn()
	if column == "" {
		return fmt.Errorf("invalid message type")
	}
	q := query.NewQuery()
	q.AddFilter(squirrel.Eq{column: true})
	q.SetLimit(Limit)
	users, err := s.store.GetUsers(ctx, q)
	if err != nil {
		return err
	}
	for _, user := range users {
		uid, err := uuid.NewV7()
		if err != nil {
			return err
		}
		sendRequest := &models.SendRequest{
			RequestID:   uid.String(),
			LineID:      user.LineID,
			RequestType: req.Typ.GetRequestType(),
		}
		profile, err := s.lineClient.GetProfile(user.LineID)
		if err != nil {
			sendRequest.Status = models.SendRequestStatusFailed
			sendRequest.Error = err.Error()
			s.store.CreateSendRequest(ctx, sendRequest)
			continue
		}
		pushMsg := &messaging_api.PushMessageRequest{
			To: user.LineID,
			Messages: []messaging_api.MessageInterface{
				&messaging_api.TextMessageV2{
					Text: fmt.Sprintf("Hi %s, %s", profile.DisplayName, req.Msg),
				},
			},
		}
		_, err = s.lineClient.PushMessage(pushMsg, uid.String())
		if err != nil {
			sendRequest.Status = models.SendRequestStatusFailed
			sendRequest.Error = err.Error()
			s.store.CreateSendRequest(ctx, sendRequest)
			continue
		}
	}
	return nil
}
