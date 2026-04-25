package v3

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/gpt"
)

func TestGetTargetSuggestion(t *testing.T) {
	key := os.Getenv("OPENAI_API_KEY")
	c := NewClient(key)

	newUser := &models.User{
		LineID:          "123",
		Height:          150,
		Weight:          40,
		TargetWeight:    50,
		TargetTimeframe: 4,
		Age:             18,
		Gender:          models.GenderFemale,
	}
	msg, _, err := c.GetTargetSuggestion(context.Background(), &gpt.BasicUserInfo{
		UserProfile: newUser.ToProfileString(),
	})
	if err != nil {
		t.Errorf("GetTargetSuggestion failed: %v", err)
	}
	fmt.Println(msg)
}
