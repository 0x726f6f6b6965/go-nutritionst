package timezone

import "time"

func GetTaipeiNow() time.Time {
	// use CST(China Standard Time)
	loc := time.FixedZone("CST", 8*3600)
	return time.Now().In(loc)
}

func GetTaipeiDate() string {
	return GetTaipeiNow().Format("2006-01-02")
}
