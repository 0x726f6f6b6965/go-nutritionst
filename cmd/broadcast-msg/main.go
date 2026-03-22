package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"go.uber.org/zap"
)

var (
	dbHost string
	dbPort string
	dbUser string
	dbPwd  string
	dbName string

	channelToken string

	msg string
)

func main() {
	logger, err := zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	initVar()

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPwd, dbHost, dbPort, dbName)

	if channelToken == "" {
		logger.Fatal("CHANNEL_ACCESS_TOKEN must be set")
	}

	// Database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Error("Unable to connect to database", zap.Error(err))
	}
	defer pool.Close()

	store := storage.NewPostgres(pool)

	client, err := messaging_api.NewMessagingApiAPI(
		channelToken,
	)
	if err != nil {
		logger.Fatal("Failed to create LINE client", zap.Error(err))
	}

	req := &messaging_api.BroadcastRequest{
		Messages: []messaging_api.MessageInterface{
			&messaging_api.TextMessageV2{
				Text: msg,
			},
		},
	}
	uid, err := uuid.NewV7()
	if err != nil {
		logger.Fatal("Failed to generate UUID", zap.Error(err))
	}

	_, err = client.Broadcast(req, uid.String())
	if err != nil {
		logger.Error("Failed to broadcast message", zap.Error(err))
		if sendErr := store.CreateSendRequest(ctx, &models.SendRequest{
			RequestID:   uid.String(),
			RequestType: models.SendRequestTypeBroadcast,
			Status:      models.SendRequestStatusFailed,
			Data:        msg,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}); sendErr != nil {
			logger.Error("Failed to create send request", zap.Error(sendErr))
		}
	}
	if err := store.CreateSendRequest(ctx, &models.SendRequest{
		RequestID:   uid.String(),
		RequestType: models.SendRequestTypeBroadcast,
		Status:      models.SendRequestStatusSuccess,
		Data:        msg,
		UpdatedAt:   time.Now(),
		CreatedAt:   time.Now(),
	}); err != nil {
		logger.Error("Failed to update send request", zap.Error(err))
	}
}

func initVar() {
	dbHost = os.Getenv("POSTGRES_HOST")
	dbPort = os.Getenv("POSTGRES_PORT")
	dbUser = os.Getenv("POSTGRES_USER")
	dbPwd = os.Getenv("POSTGRES_PASSWORD")
	dbName = os.Getenv("POSTGRES_DB")

	channelToken = os.Getenv("CHANNEL_ACCESS_TOKEN")

	msg = os.Getenv("BROADCAST_MSG")
}
