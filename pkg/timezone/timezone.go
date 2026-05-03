package timezone

import "time"

func GetTaipeiNow() time.Time {
	loc, _ := time.LoadLocation("Asia/Taipei")
	return time.Now().In(loc)
}

func GetTaipeiDate() string {
	return GetTaipeiNow().Format("2006-01-02")
}
