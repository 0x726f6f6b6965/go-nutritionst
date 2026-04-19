package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage"
	"github.com/0x726f6f6b6965/go-nutritionst/service/push"
	"github.com/0x726f6f6b6965/go-nutritionst/service/push/msg"
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

	channelSecret string
	channelToken  string

	pushMsgType msg.MsgType
	pushMsg     string
)

func main() {
	// logger
	logger, err := zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	initVar()

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPwd, dbHost, dbPort, dbName)

	if channelSecret == "" || channelToken == "" {
		logger.Fatal("CHANNEL_SECRET, CHANNEL_ACCESS_TOKEN must be set")
	}

	// Database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Error("Unable to connect to database", zap.Error(err))
	}
	defer pool.Close()

	store := storage.NewPostgres(pool)

	// line api
	client, err := messaging_api.NewMessagingApiAPI(channelToken)
	if err != nil {
		logger.Error("Unable to create line client", zap.Error(err))
		return
	}

	pushService := push.NewService(client, store)

	err = pushService.PushMsg(ctx, &push.PushMsgRequest{
		Typ: pushMsgType,
		Msg: pushMsg,
	})
	if err != nil {
		logger.Error("Unable to push message", zap.Error(err))
		return
	}
	logger.Info("Success push message", zap.String("msg_type", pushMsgType.String()), zap.String("msg", pushMsg))
}

func initVar() {
	dbHost = os.Getenv("POSTGRES_HOST")
	dbPort = os.Getenv("POSTGRES_PORT")
	dbUser = os.Getenv("POSTGRES_USER")
	dbPwd = os.Getenv("POSTGRES_PASSWORD")
	dbName = os.Getenv("POSTGRES_DB")

	channelSecret = os.Getenv("CHANNEL_SECRET")
	channelToken = os.Getenv("CHANNEL_ACCESS_TOKEN")
	// push msg type
	pushMsgType = msg.GetMsgType(os.Getenv("PUSH_MSG_TYPE"))
	pushMsg = os.Getenv("PUSH_MSG")
}
