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
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	Limit = 1000
)

type Service struct {
	lineClient *messaging_api.MessagingApiAPI
	store      *storage.Postgres
	logger     *zap.Logger
}

type PushMsgRequest struct {
	Msg string
	Typ msg.MsgType
}

func NewService(lineClient *messaging_api.MessagingApiAPI, db *storage.Postgres, logger *zap.Logger) *Service {
	return &Service{
		lineClient: lineClient,
		store:      db,
		logger:     logger,
	}
}

func (s *Service) PushMsg(ctx context.Context, req *PushMsgRequest) error {
	column := req.Typ.GetColumn()
	if column == "" {
		return fmt.Errorf("invalid message type")
	}
	keepGoing := true
	startID := 0
	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(5)
	for keepGoing {
		q := query.NewQuery()
		q.AddFilter(squirrel.Eq{column: true})
		q.AddFilter(squirrel.Gt{"id": startID})
		q.SetLimit(Limit)
		q.AddSortBy("id", false)
		users, err := s.store.GetUsers(ctx, q)
		if err != nil {
			return err
		}
		if len(users) < Limit {
			keepGoing = false
		}
		lastUser := users[len(users)-1]
		startID = lastUser.ID
		reqSendChain := make(chan *models.SendRequest, Limit)
		for _, user := range users {
			g.Go(func() error {
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
					reqSendChain <- sendRequest
					return err
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
					reqSendChain <- sendRequest
					return err
				}
				sendRequest.Status = models.SendRequestStatusSuccess
				reqSendChain <- sendRequest
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			s.logger.Error("Error:", zap.Error(err))
		}
		close(reqSendChain)
		// batch insert send request
		arr := make([]*models.SendRequest, 0, Limit)
		for sendRequest := range reqSendChain {
			arr = append(arr, sendRequest)
		}
		if err := s.store.BatchCreateSendRequests(ctx, arr); err != nil {
			s.logger.Error("Error:", zap.Error(err))
			return err
		}
	}
	return nil
}
