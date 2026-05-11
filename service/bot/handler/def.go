package handler

import (
	"context"
	"strings"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/cache"
	"github.com/0x726f6f6b6965/go-nutritionst/service/bot/ai"
	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"go.uber.org/zap"
)

var _ HandlerInterface = (*Handler)(nil)

type HandlerInterface interface {
	HandlePostback(ctx context.Context, event *linebot.Event) error
	HandleTextMessage(ctx context.Context, event *linebot.Event, message *linebot.TextMessage) error
	HandleImageMessage(ctx context.Context, event *linebot.Event, message *linebot.ImageMessage) error
}

type Handler struct {
	store         *storage.Postgres
	cache         *cache.UserContext
	lineClient    *messaging_api.MessagingApiAPI
	blobClient    *messaging_api.MessagingApiBlobAPI
	aiAPI         ai.NutritionAPI
	replyText     func(_ context.Context, replyToken, text string) error
	replyFlex     func(_ context.Context, replyToken, altText string, container messaging_api.FlexContainerInterface) error
	sendMsg       func(_ context.Context, userID string, xLineRetryKey string, msg messaging_api.MessageInterface) error
	maxDailyToken int64
	logger        *zap.Logger
}

func NewHandler(store *storage.Postgres,
	cache *cache.UserContext,
	lineClient *messaging_api.MessagingApiAPI,
	blobClient *messaging_api.MessagingApiBlobAPI,
	aiAPI ai.NutritionAPI,
	replyText func(_ context.Context, replyToken, text string) error,
	replyFlex func(_ context.Context, replyToken, altText string, container messaging_api.FlexContainerInterface) error,
	sendMsg func(_ context.Context, userID string, xLineRetryKey string, msg messaging_api.MessageInterface) error,
	maxDailyToken int64,
	logger *zap.Logger) *Handler {
	return &Handler{
		store:         store,
		cache:         cache,
		lineClient:    lineClient,
		blobClient:    blobClient,
		replyText:     replyText,
		replyFlex:     replyFlex,
		sendMsg:       sendMsg,
		aiAPI:         aiAPI,
		maxDailyToken: maxDailyToken,
		logger:        logger,
	}
}

func (h *Handler) getUserInfo(ctx context.Context, userID string) (*models.User, error) {
	user, err := h.store.GetUserByLineID(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") { // pgx v5 returns ErrNoRows equivalent
			return nil, nil // Not found
		}
		// squirrel/pgx might return different error structure, but usually pgx.ErrNoRows
		// Check if it's actually no rows
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func getMealName(m int) string {
	switch m {
	case 1:
		return "早餐"
	case 2:
		return "午餐"
	case 3:
		return "晚餐"
	case 4:
		return "點心"
	default:
		return "未知"
	}
}

type DescriptionMsg string

const (
	DescriptionMsgSetUpSuccess            DescriptionMsg = "設定成功!"
	DescriptionMsgNoProblem               DescriptionMsg = "沒問題!"
	DescriptionMsgOpen                    DescriptionMsg = "開啟"
	DescriptionMsgClose                   DescriptionMsg = "關閉"
	DescriptionMsgUpdateAlermSuccess      DescriptionMsg = "更新成功, 已將%s推播功能%s"
	DescriptionMsgAskSendWeight           DescriptionMsg = "請輸入今日體重"
	DescriptionMsgAskSendSleep            DescriptionMsg = "請輸入今日睡眠時數"
	DescriptionMsgAskSendWater            DescriptionMsg = "請輸入今日飲水量"
	DescriptionMsgAskSetNewTargetWeight   DescriptionMsg = "請輸入新的目標體重(公斤)"
	DescriptionMsgAIAnalyze               DescriptionMsg = "AI 分析中，請稍後"
	DescriptionMsgSignUpSuccess           DescriptionMsg = "註冊成功!"
	DescriptionMsgMeal                    DescriptionMsg = "餐點"
	DescriptionMsgMealName                DescriptionMsg = "名稱"
	DescriptionMsgAskSendMealName         DescriptionMsg = "請輸入%s餐點名稱"
	DescriptionMsgWeightUpdate            DescriptionMsg = "體重已更新 %.1f 公斤"
	DescriptionMsgWeightFormatError       DescriptionMsg = "體重格式錯誤，請重新輸入 ex. 70.5"
	DescriptionMsgSleepUpdate             DescriptionMsg = "睡眠時數已更新 %.1f 小時"
	DescriptionMsgSleepFormatError        DescriptionMsg = "睡眠時數格式錯誤，請重新輸入 ex. 7.5"
	DescriptionMsgWaterUpdate             DescriptionMsg = "飲水量已更新 %.1f 毫升"
	DescriptionMsgWaterFormatError        DescriptionMsg = "飲水量格式錯誤，請重新輸入 ex. 2000"
	DescriptionMsgTargetWeightUpdate      DescriptionMsg = "目標已更新為 %.1f 公斤，預計 %d 個月達成!"
	DescriptionMsgAskSetTargetWeight      DescriptionMsg = "請輸入目標體重 ex. 65.0"
	DescriptionMsgTargetTimeFormatError   DescriptionMsg = "目標時間格式錯誤，請重新輸入 ex. 3"
	DescriptionMsgAskSetTargetTime        DescriptionMsg = "請輸入目標時間(月) ex. 3"
	DescriptionMsgTargetWeightFormatError DescriptionMsg = "目標體重格式錯誤，請重新輸入 ex. 65.0"
	DescriptionMsgGender                  DescriptionMsg = "性別"
	DescriptionMsgMale                    DescriptionMsg = "男"
	DescriptionMsgFemale                  DescriptionMsg = "女"
	DescriptionMsgHeight                  DescriptionMsg = "身高"
	DescriptionMsgWeight                  DescriptionMsg = "體重"
	DescriptionMsgAge                     DescriptionMsg = "年齡"
	DescriptionMsgTargetWeight            DescriptionMsg = "目標體重"
	DescriptionMsgTargetTime              DescriptionMsg = "目標時間"
	DescriptionMsgBasicInfo               DescriptionMsg = "基本資料"
	DescriptionMsgAskSetGender            DescriptionMsg = "請輸入性別(男/女)"
	DescriptionMsgAskHeight               DescriptionMsg = "請輸入身高(公分) ex. 175.5"
	DescriptionMsgAskWeight               DescriptionMsg = "請輸入體重(公斤) ex. 70.5"
	DescriptionMsgAskAge                  DescriptionMsg = "請輸入年齡 ex. 25"
	DescriptionMsgGenderFormatError       DescriptionMsg = "性別格式錯誤，請重新輸入(男/女)"
	DescriptionMsgHeightFormatError       DescriptionMsg = "身高格式錯誤，請重新輸入 ex. 175.5"
	DescriptionMsgAgeFormatError          DescriptionMsg = "年齡格式錯誤，請重新輸入 ex. 25"
	DescriptionMsgWelcomeSignUp           DescriptionMsg = "歡迎使用營養師機器人，請依序輸入您的身高(公分)、體重(公斤)、年齡、性別(男/女)以完成註冊"
	DescriptionMsgUploadMeal              DescriptionMsg = "上傳餐點"
	DescriptionMsgUploadMealIntro         DescriptionMsg = "請點擊下方選單開始上傳餐點"
)

func (m DescriptionMsg) String() string {
	return string(m)
}
