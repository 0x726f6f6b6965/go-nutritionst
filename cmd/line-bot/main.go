package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage"
	v2 "github.com/0x726f6f6b6965/go-nutritionst/pkg/gpt/v2"
	"github.com/0x726f6f6b6965/go-nutritionst/service/bot"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/line/line-bot-sdk-go/v8/linebot"
	"go.uber.org/zap"
)

func main() {
	// logger
	logger, err := zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPwd := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPwd, dbHost, dbPort, dbName)

	channelSecret := os.Getenv("CHANNEL_SECRET")
	channelToken := os.Getenv("CHANNEL_ACCESS_TOKEN")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	if channelSecret == "" || channelToken == "" || openaiKey == "" {
		logger.Fatal("CHANNEL_SECRET, CHANNEL_ACCESS_TOKEN, OPENAI_API_KEY must be set")
	}

	// Database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Error("Unable to connect to database", zap.Error(err))
	}
	defer pool.Close()

	store := storage.NewPostgres(pool)

	// GPT
	gptClient := v2.NewClient(openaiKey)

	// Bot Service
	botService, err := bot.NewService(channelToken, store, gptClient,
		bot.WithLogger(logger),
		bot.WithMaxDailyToken(1000000))
	if err != nil {
		logger.Error("Failed to initialize bot service", zap.Error(err))
	}

	// Line Handler (Legacy for parsing)
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		events, err := linebot.ParseRequest(channelSecret, r)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}

		for _, event := range events {
			if err := botService.HandleEvent(ctx, event); err != nil {
				logger.Error("HandleEvent error", zap.Error(err))
			}
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Listening on :%s", zap.String("port", port))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatal("ListenAndServe error", zap.Error(err))
	}
}
