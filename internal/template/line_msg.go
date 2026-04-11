package template

import (
	"fmt"
	"time"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

type Variable struct {
	Name  string
	Value string
}

func GetCheckMsg(title string, vars []Variable, actions []string) *messaging_api.FlexBubble {
	contents := []messaging_api.FlexComponentInterface{}

	for _, v := range vars {
		item := &messaging_api.FlexBox{
			Layout:  messaging_api.FlexBoxLAYOUT_BASELINE,
			Spacing: "sm",
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexText{
					Text:  v.Name,
					Color: "#aaaaaa",
					Size:  "sm",
					Flex:  2,
				},
				&messaging_api.FlexText{
					Text:  v.Value,
					Wrap:  true,
					Color: "#666666",
					Size:  "sm",
					Flex:  5,
				},
			},
		}
		contents = append(contents, item)
	}

	return &messaging_api.FlexBubble{
		Hero: &messaging_api.FlexImage{
			Url:         "https://developers-resource.landpress.line.me/fx/img/01_1_cafe.png",
			Size:        "full",
			AspectRatio: "20:13",
			AspectMode:  messaging_api.FlexImageASPECT_MODE_COVER,
			Action: &messaging_api.UriAction{
				Uri: "https://line.me/",
			},
		},
		Body: &messaging_api.FlexBox{
			Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexText{
					Text:   title,
					Weight: messaging_api.FlexTextWEIGHT_BOLD,         // Enum
					Size:   string(messaging_api.FlexTextFontSize_XL), // String
				},
				&messaging_api.FlexBox{
					Layout:   messaging_api.FlexBoxLAYOUT_VERTICAL,
					Margin:   "lg",
					Spacing:  "sm",
					Contents: contents,
				},
			},
		},
		Footer: &messaging_api.FlexBox{
			Layout:  messaging_api.FlexBoxLAYOUT_VERTICAL,
			Spacing: "sm",
			Flex:    0,
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexButton{
					Height: messaging_api.FlexButtonHEIGHT_SM,
					Action: &messaging_api.PostbackAction{
						Label: "正確",
						Data:  actions[0],
					},
				},
				&messaging_api.FlexButton{
					Style:  messaging_api.FlexButtonSTYLE_LINK,
					Height: messaging_api.FlexButtonHEIGHT_SM,
					Action: &messaging_api.PostbackAction{
						Label: "錯誤",
						Data:  actions[1],
					},
				},
			},
		},
	}
}

func GetUploadMsg(vars []Variable) *messaging_api.FlexBubble {
	contents := []messaging_api.FlexComponentInterface{}

	for _, v := range vars {
		item := &messaging_api.FlexBox{
			Layout:  messaging_api.FlexBoxLAYOUT_BASELINE,
			Spacing: "sm",
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexText{
					Text:  v.Name,
					Color: "#aaaaaa",
					Size:  "sm",
					Flex:  1,
				},
				&messaging_api.FlexText{
					Text:  v.Value,
					Wrap:  true,
					Color: "#666666",
					Size:  "sm",
					Flex:  5,
				},
			},
		}
		contents = append(contents, item)
	}

	return &messaging_api.FlexBubble{
		Hero: &messaging_api.FlexImage{
			Url:         "https://developers-resource.landpress.line.me/fx/img/01_1_cafe.png",
			Size:        "full",
			AspectRatio: "20:13",
			AspectMode:  messaging_api.FlexImageASPECT_MODE_COVER,
			Action: &messaging_api.UriAction{
				Uri: "https://line.me/",
			},
		},
		Body: &messaging_api.FlexBox{
			Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexText{
					Text:   "上傳餐點",
					Weight: messaging_api.FlexTextWEIGHT_BOLD,
					Size:   string(messaging_api.FlexTextFontSize_XL),
				},
				&messaging_api.FlexBox{
					Layout:   messaging_api.FlexBoxLAYOUT_VERTICAL,
					Margin:   "lg",
					Spacing:  "sm",
					Contents: contents,
				},
			},
		},
		Footer: &messaging_api.FlexBox{
			Layout:  messaging_api.FlexBoxLAYOUT_VERTICAL,
			Spacing: "sm",
			Flex:    0,
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexButton{
					Style:  messaging_api.FlexButtonSTYLE_LINK,
					Height: messaging_api.FlexButtonHEIGHT_SM,
					Action: &messaging_api.UriAction{
						Label: "開啟相機",
						Uri:   "https://line.me/R/nv/camera/",
					},
				},
				&messaging_api.FlexButton{
					Style:  messaging_api.FlexButtonSTYLE_LINK,
					Height: messaging_api.FlexButtonHEIGHT_SM,
					Action: &messaging_api.UriAction{
						Label: "選擇圖片",
						Uri:   "https://line.me/R/nv/cameraRoll/single",
					},
				},
			},
		},
	}
}

func GetSetMealMsg() *messaging_api.FlexBubble {
	meals := []struct {
		Label string
		Value string
	}{
		{"早餐", "1"}, {"午餐", "2"}, {"晚餐", "3"}, {"點心", "4"},
	}

	buttons := []messaging_api.FlexComponentInterface{}
	for _, m := range meals {
		buttons = append(buttons, &messaging_api.FlexButton{
			Style:  messaging_api.FlexButtonSTYLE_LINK,
			Height: messaging_api.FlexButtonHEIGHT_SM,
			Action: &messaging_api.PostbackAction{
				Label: m.Label,
				Data:  fmt.Sprintf("action=set_meal&data=%s", m.Value),
			},
		})
	}

	return &messaging_api.FlexBubble{
		Body: &messaging_api.FlexBox{
			Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexText{
					Text:   "設定餐點",
					Weight: messaging_api.FlexTextWEIGHT_BOLD,
					Size:   string(messaging_api.FlexTextFontSize_XL),
				},
			},
		},
		Footer: &messaging_api.FlexBox{
			Layout:   messaging_api.FlexBoxLAYOUT_VERTICAL,
			Spacing:  "sm",
			Contents: buttons,
			Flex:     0,
		},
	}
}

func GetMonthMsg(inputDate, startDate time.Time, act string) *messaging_api.FlexBubble {
	today := time.Now()
	// Normalize today to start of day for comparison
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())

	year, month, _ := inputDate.Date()

	header := "設定結束日期"
	if act == "set_report_start" {
		header = "設定開始日期"
	}

	// Days Header
	daysHeader := &messaging_api.FlexBox{
		Layout:   messaging_api.FlexBoxLAYOUT_HORIZONTAL,
		Margin:   "lg",
		Spacing:  "sm",
		Contents: []messaging_api.FlexComponentInterface{},
	}
	days := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	for _, d := range days {
		daysHeader.Contents = append(daysHeader.Contents, &messaging_api.FlexText{
			Text:  d,
			Color: "#aaaaaa",
			Size:  "sm",
			Flex:  1,
			Align: messaging_api.FlexTextALIGN_CENTER,
		})
	}

	// Weeks
	firstDayOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, inputDate.Location())
	// lastDayOfMonth via adding one month then subtracting a day
	firstDayOfNextMonth := firstDayOfMonth.AddDate(0, 1, 0)
	lastDayOfMonth := firstDayOfNextMonth.AddDate(0, 0, -1)
	numDays := lastDayOfMonth.Day()

	padding := int(firstDayOfMonth.Weekday()) // Sun=0, Mon=1, etc.

	// Rows of days
	rows := []messaging_api.FlexComponentInterface{}
	currentRow := &messaging_api.FlexBox{
		Layout:   messaging_api.FlexBoxLAYOUT_HORIZONTAL,
		Margin:   "md",
		Contents: []messaging_api.FlexComponentInterface{},
	}

	// Add Padding
	for i := 0; i < padding; i++ {
		currentRow.Contents = append(currentRow.Contents, &messaging_api.FlexText{
			Text:  " ",
			Size:  "sm",
			Flex:  1,
			Align: messaging_api.FlexTextALIGN_CENTER,
		})
	}

	colCount := padding
	for i := 1; i <= numDays; i++ {
		if colCount == 7 {
			rows = append(rows, currentRow)
			currentRow = &messaging_api.FlexBox{
				Layout:   messaging_api.FlexBoxLAYOUT_HORIZONTAL,
				Margin:   "md",
				Contents: []messaging_api.FlexComponentInterface{},
			}
			colCount = 0
		}

		currDate := time.Date(year, month, i, 0, 0, 0, 0, inputDate.Location())

		isTodayStart := currDate.Equal(today) && today.Equal(startDate)
		isSelectedStart := currDate.Equal(startDate) && !today.Equal(startDate)
		isFuture := currDate.After(today)
		isWayPast := today.Sub(currDate) > 90*24*time.Hour

		if isTodayStart || isSelectedStart {
			currentRow.Contents = append(currentRow.Contents, &messaging_api.FlexBox{
				Layout:         messaging_api.FlexBoxLAYOUT_VERTICAL,
				Flex:           1,
				AlignItems:     messaging_api.FlexBoxALIGN_ITEMS_CENTER,     // Enum
				JustifyContent: messaging_api.FlexBoxJUSTIFY_CONTENT_CENTER, // Enum
				Contents: []messaging_api.FlexComponentInterface{
					&messaging_api.FlexBox{
						Layout:          messaging_api.FlexBoxLAYOUT_VERTICAL,
						Height:          "30px",
						Width:           "30px",
						CornerRadius:    "100px",
						BackgroundColor: "#00B900",
						AlignItems:      messaging_api.FlexBoxALIGN_ITEMS_CENTER,     // Enum
						JustifyContent:  messaging_api.FlexBoxJUSTIFY_CONTENT_CENTER, // Enum
						Contents: []messaging_api.FlexComponentInterface{
							&messaging_api.FlexText{
								Text:    fmt.Sprintf("%d", i),
								Size:    "sm",
								Color:   "#ffffff",
								Align:   messaging_api.FlexTextALIGN_CENTER,
								Gravity: messaging_api.FlexTextGRAVITY_CENTER,
								Weight:  messaging_api.FlexTextWEIGHT_BOLD,
								Action: &messaging_api.PostbackAction{
									Label: "set date",
									Data:  fmt.Sprintf("action=%s&data=%s", act, currDate.Format("2006-01-02")),
								},
							},
						},
					},
				},
			})
		} else if isFuture || isWayPast {
			currentRow.Contents = append(currentRow.Contents, &messaging_api.FlexText{
				Text:    fmt.Sprintf("%d", i),
				Size:    "sm",
				Flex:    1,
				Align:   messaging_api.FlexTextALIGN_CENTER,
				Gravity: messaging_api.FlexTextGRAVITY_CENTER,
				Color:   "#C6C6C6",
			})
		} else {
			currentRow.Contents = append(currentRow.Contents, &messaging_api.FlexText{
				Text:    fmt.Sprintf("%d", i),
				Size:    "sm",
				Flex:    1,
				Align:   messaging_api.FlexTextALIGN_CENTER,
				Gravity: messaging_api.FlexTextGRAVITY_CENTER,
				Action: &messaging_api.PostbackAction{
					Label: "set date",
					Data:  fmt.Sprintf("action=%s&data=%s", act, currDate.Format("2006-01-02")),
				},
			})
		}
		colCount++
	}
	// Fill remaining row
	for colCount < 7 {
		currentRow.Contents = append(currentRow.Contents, &messaging_api.FlexText{
			Text:  " ",
			Size:  "sm",
			Flex:  1,
			Align: messaging_api.FlexTextALIGN_CENTER,
		})
		colCount++
	}
	rows = append(rows, currentRow)

	bodyContents := []messaging_api.FlexComponentInterface{
		&messaging_api.FlexText{
			Text:   header,
			Weight: messaging_api.FlexTextWEIGHT_BOLD,
			Size:   string(messaging_api.FlexTextFontSize_XL),
			Align:  messaging_api.FlexTextALIGN_CENTER,
			Margin: "md",
		},
		&messaging_api.FlexText{
			Text:   fmt.Sprintf("%d 年 %d 月", year, month),
			Weight: messaging_api.FlexTextWEIGHT_BOLD,
			Size:   string(messaging_api.FlexTextFontSize_XL),
			Align:  messaging_api.FlexTextALIGN_CENTER,
			Margin: "md",
		},
		&messaging_api.FlexSeparator{
			Margin: "lg",
		},
		daysHeader,
	}
	bodyContents = append(bodyContents, rows...)

	// Footer logic
	footerContents := []messaging_api.FlexComponentInterface{}

	// Prev Month Button
	oneMonthAgo := time.Date(year, month-1, 1, 0, 0, 0, 0, inputDate.Location())
	if today.Sub(inputDate) <= 90*24*time.Hour {
		prevDate := oneMonthAgo
		footerContents = append(footerContents, &messaging_api.FlexButton{
			Style:  messaging_api.FlexButtonSTYLE_LINK,
			Height: messaging_api.FlexButtonHEIGHT_SM,
			Action: &messaging_api.PostbackAction{
				Label: "上個月",
				Data:  fmt.Sprintf("action=get_month&data=%s&data=%s", prevDate.Format("2006-01-02"), act),
			},
		})
	} else {
		footerContents = append(footerContents, &messaging_api.FlexText{
			Text:  " ",
			Size:  "sm",
			Flex:  1,
			Align: messaging_api.FlexTextALIGN_CENTER,
		})
	}

	// Next Month Button
	firstDayOfCurrentMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	if firstDayOfMonth.Before(firstDayOfCurrentMonth) {
		nextDate := firstDayOfMonth.AddDate(0, 1, 0)
		footerContents = append(footerContents, &messaging_api.FlexButton{
			Style:  messaging_api.FlexButtonSTYLE_LINK,
			Height: messaging_api.FlexButtonHEIGHT_SM,
			Action: &messaging_api.PostbackAction{
				Label: "下個月",
				Data:  fmt.Sprintf("action=get_month&data=%s&data=%s", nextDate.Format("2006-01-02"), act),
			},
		})
	} else {
		footerContents = append(footerContents, &messaging_api.FlexText{
			Text:  " ",
			Size:  "sm",
			Flex:  1,
			Align: messaging_api.FlexTextALIGN_CENTER,
		})
	}

	return &messaging_api.FlexBubble{
		Body: &messaging_api.FlexBox{
			Layout:   messaging_api.FlexBoxLAYOUT_VERTICAL,
			Spacing:  "md",
			Contents: bodyContents,
		},
		Footer: &messaging_api.FlexBox{
			Layout:   messaging_api.FlexBoxLAYOUT_HORIZONTAL,
			Spacing:  "sm",
			Flex:     1,
			Contents: footerContents,
		},
	}
}

func GetAIMealResponseMsg(response *models.MealHistory, cost float64) *messaging_api.FlexBubble {
	return &messaging_api.FlexBubble{
		Body: &messaging_api.FlexBox{
			Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexText{
					Text:   "營養評估與建議",
					Weight: messaging_api.FlexTextWEIGHT_BOLD,
					Color:  "#1DB446",
					Size:   "sm",
				},
				&messaging_api.FlexText{
					Text:   response.Description,
					Weight: messaging_api.FlexTextWEIGHT_BOLD,
					Size:   string(messaging_api.FlexTextFontSize_LG),
					Margin: "md",
					Wrap:   true,
				},
				&messaging_api.FlexSeparator{
					Margin: "xxl",
				},
				&messaging_api.FlexBox{
					Layout:  messaging_api.FlexBoxLAYOUT_VERTICAL,
					Margin:  "xxl",
					Spacing: "sm",
					Contents: []messaging_api.FlexComponentInterface{
						nutritionRow("卡路里(kcal)", response.CaloriesKcal),
						nutritionRow("蛋白質(g)", response.ProteinG),
						nutritionRow("碳水化合物(g)", response.CarbsG),
						nutritionRow("脂肪(g)", response.FatG),
						nutritionRow("鈉(mg)", response.SodiumMg),
						&messaging_api.FlexSeparator{
							Margin: "xxl",
						},
						&messaging_api.FlexBox{
							Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
							Contents: []messaging_api.FlexComponentInterface{
								&messaging_api.FlexBox{
									Layout: messaging_api.FlexBoxLAYOUT_HORIZONTAL,
									Contents: []messaging_api.FlexComponentInterface{
										&messaging_api.FlexText{
											Text:   "簡單描述",
											Size:   string(messaging_api.FlexTextFontSize_MD),
											Flex:   0,
											Margin: "none",
											Weight: messaging_api.FlexTextWEIGHT_BOLD,
											Color:  "#1DB446",
										},
									},
								},
								&messaging_api.FlexBox{
									Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
									Contents: []messaging_api.FlexComponentInterface{
										&messaging_api.FlexText{
											Text:   strOrNone(response.AIDescription),
											Margin: "md",
											Size:   "sm",
											Wrap:   true,
										},
									},
								},
							},
						},
						&messaging_api.FlexSeparator{
							Margin: "xxl",
						},
						&messaging_api.FlexBox{
							Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
							Contents: []messaging_api.FlexComponentInterface{
								&messaging_api.FlexBox{
									Layout: messaging_api.FlexBoxLAYOUT_HORIZONTAL,
									Contents: []messaging_api.FlexComponentInterface{
										&messaging_api.FlexText{
											Text:   "建議",
											Size:   string(messaging_api.FlexTextFontSize_MD),
											Flex:   0,
											Margin: "none",
											Weight: messaging_api.FlexTextWEIGHT_BOLD,
											Color:  "#1DB446",
										},
									},
								},
								&messaging_api.FlexBox{
									Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
									Contents: []messaging_api.FlexComponentInterface{
										&messaging_api.FlexText{
											Text:   strOrNone(response.AISuggest),
											Margin: "md",
											Size:   "sm",
											Wrap:   true,
										},
									},
								},
							},
						},
					},
				},
				&messaging_api.FlexSeparator{
					Margin: "xxl",
				},
				&messaging_api.FlexBox{
					Layout: messaging_api.FlexBoxLAYOUT_HORIZONTAL,
					Margin: "md",
					Contents: []messaging_api.FlexComponentInterface{
						&messaging_api.FlexText{
							Text:  "分析時間(s)",
							Size:  string(messaging_api.FlexTextFontSize_XS),
							Color: "#aaaaaa",
							Flex:  0,
						},
						&messaging_api.FlexText{
							Text:  fmt.Sprintf("%.2f s", cost),
							Color: "#aaaaaa",
							Size:  string(messaging_api.FlexTextFontSize_XS),
							Align: messaging_api.FlexTextALIGN_END,
						},
					},
				},
			},
		},
		Styles: &messaging_api.FlexBubbleStyles{
			Footer: &messaging_api.FlexBlockStyle{
				Separator: true,
			},
		},
	}
}

func nutritionRow(label string, value float64) *messaging_api.FlexBox {
	return &messaging_api.FlexBox{
		Layout: messaging_api.FlexBoxLAYOUT_HORIZONTAL,
		Contents: []messaging_api.FlexComponentInterface{
			&messaging_api.FlexText{
				Text:  label,
				Size:  "sm",
				Color: "#555555",
				Flex:  0,
			},
			&messaging_api.FlexText{
				Text:  fmt.Sprintf("%.2f", value),
				Size:  "sm",
				Color: "#111111",
				Flex:  1,
				Align: messaging_api.FlexTextALIGN_END,
				Wrap:  true,
			},
		},
	}
}

func nutritionRowWithObserve(label string, value float64, observe string, needed float64, unit string) *messaging_api.FlexBox {
	prefix := "多餘"
	if needed < 0 {
		prefix = "還需"
		needed = -needed
	}

	return &messaging_api.FlexBox{
		Layout: messaging_api.FlexBoxLAYOUT_HORIZONTAL,
		Contents: []messaging_api.FlexComponentInterface{
			&messaging_api.FlexText{
				Text:  label,
				Size:  "sm",
				Color: "#555555",
				Flex:  0,
			},
			&messaging_api.FlexText{
				Text:  fmt.Sprintf("%.2f; %s; %s%.2f%s", value, observe, prefix, needed, unit),
				Size:  "sm",
				Wrap:  true,
				Flex:  1,
				Color: "#111111",
				Align: messaging_api.FlexTextALIGN_END,
			},
		},
	}
}

func strOrNone(s ...string) string {
	if len(s) == 0 {
		return "無"
	}
	if len(s) == 1 {
		return s[0]
	}
	result := ""
	idx := 1
	for _, v := range s {
		if v == "" {
			continue
		}
		result += fmt.Sprintf("%d. %s\n", idx, v)
		idx++
	}
	if result == "" {
		return "無"
	}
	return result
}

func GetAIMealDailyMsg(response *models.MealDaily, cost float64) *messaging_api.FlexBubble {
	return &messaging_api.FlexBubble{
		Body: &messaging_api.FlexBox{
			Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexText{
					Text:   "營養評估與建議",
					Weight: messaging_api.FlexTextWEIGHT_BOLD,
					Color:  "#1DB446",
					Size:   "sm",
				},
				&messaging_api.FlexText{
					Text:   models.GetMealsDescription(response.Meals) + "的營養評估",
					Weight: messaging_api.FlexTextWEIGHT_BOLD,
					Size:   string(messaging_api.FlexTextFontSize_LG),
					Margin: "md",
					Wrap:   true,
				},
				&messaging_api.FlexSeparator{
					Margin: "xxl",
				},
				&messaging_api.FlexBox{
					Layout:  messaging_api.FlexBoxLAYOUT_VERTICAL,
					Margin:  "xxl",
					Spacing: "sm",
					Contents: []messaging_api.FlexComponentInterface{
						nutritionRowWithObserve("總卡路里(kcal)", response.TotalCaloriesKcal, response.ComplianceCalories, response.ComplianceDeltasCaloriesKcal, "kcal"),
						nutritionRowWithObserve("總蛋白質(g)", response.TotalProteinG, response.ComplianceProtein, response.ComplianceDeltasProteinG, "g"),
						nutritionRow("總碳水化合物(g)", response.TotalCarbsG),
						nutritionRow("總脂肪(g)", response.TotalFatG),
						nutritionRowWithObserve("總鈉(mg)", response.TotalSodiumMg, response.ComplianceSodium, response.ComplianceDeltasSodiumMg, "mg"),
						&messaging_api.FlexSeparator{
							Margin: "xxl",
						},
						&messaging_api.FlexBox{
							Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
							Contents: []messaging_api.FlexComponentInterface{
								&messaging_api.FlexBox{
									Layout: messaging_api.FlexBoxLAYOUT_HORIZONTAL,
									Contents: []messaging_api.FlexComponentInterface{
										&messaging_api.FlexText{
											Text:   "簡單描述",
											Size:   string(messaging_api.FlexTextFontSize_MD),
											Flex:   0,
											Margin: "none",
											Weight: messaging_api.FlexTextWEIGHT_BOLD,
											Color:  "#1DB446",
										},
									},
								},
								&messaging_api.FlexBox{
									Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
									Contents: []messaging_api.FlexComponentInterface{
										&messaging_api.FlexText{
											Text:   strOrNone(response.Insights...),
											Margin: "md",
											Size:   "sm",
											Wrap:   true,
										},
									},
								},
							},
						},
						&messaging_api.FlexSeparator{
							Margin: "xxl",
						},
						&messaging_api.FlexBox{
							Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
							Contents: []messaging_api.FlexComponentInterface{
								&messaging_api.FlexBox{
									Layout: messaging_api.FlexBoxLAYOUT_HORIZONTAL,
									Contents: []messaging_api.FlexComponentInterface{
										&messaging_api.FlexText{
											Text:   "建議",
											Size:   string(messaging_api.FlexTextFontSize_MD),
											Flex:   0,
											Margin: "none",
											Weight: messaging_api.FlexTextWEIGHT_BOLD,
											Color:  "#1DB446",
										},
									},
								},
								&messaging_api.FlexBox{
									Layout: messaging_api.FlexBoxLAYOUT_VERTICAL,
									Contents: []messaging_api.FlexComponentInterface{
										&messaging_api.FlexText{
											Text:   strOrNone(response.TodayCoaching...),
											Margin: "md",
											Size:   "sm",
											Wrap:   true,
										},
									},
								},
							},
						},
					},
				},
				&messaging_api.FlexSeparator{
					Margin: "xxl",
				},
				&messaging_api.FlexBox{
					Layout: messaging_api.FlexBoxLAYOUT_HORIZONTAL,
					Margin: "md",
					Contents: []messaging_api.FlexComponentInterface{
						&messaging_api.FlexText{
							Text:  "分析時間(s)",
							Size:  string(messaging_api.FlexTextFontSize_XS),
							Color: "#aaaaaa",
							Flex:  0,
						},
						&messaging_api.FlexText{
							Text:  fmt.Sprintf("%.2f s", cost),
							Color: "#aaaaaa",
							Size:  string(messaging_api.FlexTextFontSize_XS),
							Align: messaging_api.FlexTextALIGN_END,
						},
					},
				},
			},
		},
		Styles: &messaging_api.FlexBubbleStyles{
			Footer: &messaging_api.FlexBlockStyle{
				Separator: true,
			},
		},
	}
}
