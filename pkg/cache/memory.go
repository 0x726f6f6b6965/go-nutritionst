package cache

import (
	"fmt"
	"time"

	"github.com/0x726f6f6b6965/go-nutritionst/service/bot/action"
	"github.com/patrickmn/go-cache"
)

type UserContext struct {
	c *cache.Cache
}

func NewUserContext() *UserContext {
	return &UserContext{
		c: cache.New(50*time.Minute, 60*time.Minute), // Longer TTL to match python's 3000s (~50min)
	}
}

// Generic getters/setters/deleters could be useful, but specific ones match the python API better for porting

// Meal Description
func (u *UserContext) GetMealDescription(userID string) string {
	if val, found := u.c.Get(userID); found {
		return val.(string)
	}
	return ""
}

func (u *UserContext) SetMealDescription(userID string, description string) {
	u.c.Set(userID, description, cache.DefaultExpiration)
}

func (u *UserContext) DeleteMealDescription(userID string) {
	u.c.Delete(userID)
}

// Meal Type
func (u *UserContext) GetMeal(userID string) int {
	if val, found := u.c.Get(fmt.Sprintf("%s@meal", userID)); found {
		return val.(int)
	}
	return 0 // meal.Unknown
}

func (u *UserContext) SetMeal(userID string, meal int) {
	u.c.Set(fmt.Sprintf("%s@meal", userID), meal, cache.DefaultExpiration)
}

func (u *UserContext) DeleteMeal(userID string) {
	u.c.Delete(fmt.Sprintf("%s@meal", userID))
}

// Picture
func (u *UserContext) GetPicture(userID string) []byte {
	if val, found := u.c.Get(fmt.Sprintf("%s@pic", userID)); found {
		return val.([]byte)
	}
	return nil
}

func (u *UserContext) SetPicture(userID string, picture []byte) {
	u.c.Set(fmt.Sprintf("%s@pic", userID), picture, cache.DefaultExpiration)
}

func (u *UserContext) DeletePicture(userID string) {
	u.c.Delete(fmt.Sprintf("%s@pic", userID))
}

// User Profile Registration Flow
func (u *UserContext) GetRegisterProcess(userID string) bool {
	if val, found := u.c.Get(fmt.Sprintf("%s@registerProcess", userID)); found {
		return val.(bool)
	}
	return false
}

func (u *UserContext) GetChangeTargetWeightProcess(userID string) bool {
	if val, found := u.c.Get(fmt.Sprintf("%s@changeTargetWeightProcess", userID)); found {
		return val.(bool)
	}
	return false
}

func (u *UserContext) SetChangeTargetWeightProcess(userID string, isProcess bool) {
	u.c.Set(fmt.Sprintf("%s@changeTargetWeightProcess", userID), isProcess, cache.DefaultExpiration)
}

func (u *UserContext) SetRegisterProcess(userID string, isProcess bool) {
	u.c.Set(fmt.Sprintf("%s@registerProcess", userID), isProcess, cache.DefaultExpiration)
}

func (u *UserContext) DeleteRegisterProcess(userID string) {
	u.c.Delete(fmt.Sprintf("%s@registerProcess", userID))
}

// Height
func (u *UserContext) GetHeight(userID string) float64 {
	if val, found := u.c.Get(fmt.Sprintf("%s@height", userID)); found {
		return val.(float64)
	}
	return 0
}

func (u *UserContext) SetHeight(userID string, height float64) {
	u.c.Set(fmt.Sprintf("%s@height", userID), height, cache.DefaultExpiration)
}

func (u *UserContext) DeleteHeight(userID string) {
	u.c.Delete(fmt.Sprintf("%s@height", userID))
}

// Weight
func (u *UserContext) GetWeight(userID string) float64 {
	if val, found := u.c.Get(fmt.Sprintf("%s@weight", userID)); found {
		return val.(float64)
	}
	return 0
}

func (u *UserContext) SetWeight(userID string, weight float64) {
	u.c.Set(fmt.Sprintf("%s@weight", userID), weight, cache.DefaultExpiration)
}

func (u *UserContext) DeleteWeight(userID string) {
	u.c.Delete(fmt.Sprintf("%s@weight", userID))
}

// Target Weight
func (u *UserContext) GetTargetWeight(userID string) float64 {
	if val, found := u.c.Get(fmt.Sprintf("%s@targetWeight", userID)); found {
		return val.(float64)
	}
	return 0
}

func (u *UserContext) SetTargetWeight(userID string, targetWeight float64) {
	u.c.Set(fmt.Sprintf("%s@targetWeight", userID), targetWeight, cache.DefaultExpiration)
}

func (u *UserContext) DeleteTargetWeight(userID string) {
	u.c.Delete(fmt.Sprintf("%s@targetWeight", userID))
}

// Target Timeframe
func (u *UserContext) GetTargetTimeframe(userID string) int {
	if val, found := u.c.Get(fmt.Sprintf("%s@targetTimeframe", userID)); found {
		return val.(int)
	}
	return 0
}

func (u *UserContext) SetTargetTimeframe(userID string, targetTimeframe int) {
	u.c.Set(fmt.Sprintf("%s@targetTimeframe", userID), targetTimeframe, cache.DefaultExpiration)
}

func (u *UserContext) DeleteTargetTimeframe(userID string) {
	u.c.Delete(fmt.Sprintf("%s@targetTimeframe", userID))
}

// Age
func (u *UserContext) GetAge(userID string) int {
	if val, found := u.c.Get(fmt.Sprintf("%s@age", userID)); found {
		return val.(int)
	}
	return 0
}

func (u *UserContext) SetAge(userID string, age int) {
	u.c.Set(fmt.Sprintf("%s@age", userID), age, cache.DefaultExpiration)
}

func (u *UserContext) DeleteAge(userID string) {
	u.c.Delete(fmt.Sprintf("%s@age", userID))
}

// Gender
func (u *UserContext) GetGender(userID string) int {
	if val, found := u.c.Get(fmt.Sprintf("%s@gender", userID)); found {
		return val.(int)
	}
	return 0 // Unknown
}

func (u *UserContext) SetGender(userID string, gender int) {
	u.c.Set(fmt.Sprintf("%s@gender", userID), gender, cache.DefaultExpiration)
}

func (u *UserContext) DeleteGender(userID string) {
	u.c.Delete(fmt.Sprintf("%s@gender", userID))
}

// Report
func (u *UserContext) GetStartReport(userID string) string {
	if val, found := u.c.Get(fmt.Sprintf("%s@startReport", userID)); found {
		return val.(string)
	}
	return ""
}

func (u *UserContext) SetStartReport(userID string, dateStr string) {
	u.c.Set(fmt.Sprintf("%s@startReport", userID), dateStr, cache.DefaultExpiration)
}

func (u *UserContext) DeleteStartReport(userID string) {
	u.c.Delete(fmt.Sprintf("%s@startReport", userID))
}

func (u *UserContext) GetTextMessageActionType(userID string) action.TextMessageActionType {
	if val, found := u.c.Get(fmt.Sprintf("%s@textMessageActionType", userID)); found {
		return val.(action.TextMessageActionType)
	}
	return action.TextMessageActionTypeUnknown
}

func (u *UserContext) SetTextMessageActionType(userID string, actionType action.TextMessageActionType) {
	u.c.Set(fmt.Sprintf("%s@textMessageActionType", userID), actionType, cache.DefaultExpiration)
}

func (u *UserContext) DeleteTextMessageActionType(userID string) {
	u.c.Delete(fmt.Sprintf("%s@textMessageActionType", userID))
}
