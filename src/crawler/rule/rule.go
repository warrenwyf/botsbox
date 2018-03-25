package rule

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"../../common/util"
)

type Rule struct {
	json gjson.Result
}

func NewRuleWithContent(ruleContent string) (*Rule, error) {
	if !gjson.Valid(ruleContent) {
		return nil, errors.New("Invalid rule content")
	}

	r := &Rule{
		json: gjson.Parse(ruleContent),
	}

	return r, nil
}

func (self *Rule) GetInterval() (time.Duration, error) {
	defaultValue := time.Hour

	// Parse $every
	everyElem := self.json.Get("$every") // 5s | 10m | 1h
	if everyElem.Exists() {
		str := strings.ToLower(everyElem.String())
		duration, err := util.ParseDuration(str)
		if err != nil {
			return defaultValue, errors.New("Parse $every error")
		}

		return duration, nil
	}

	return defaultValue, nil
}

func (self *Rule) GetDelay() (time.Duration, error) {
	defaultValue := time.Duration(0)

	now := time.Now()

	// Parse $startDay and $startDayTime
	nextTime := now.Add(0)
	startDayElem := self.json.Get("$startDay") // w0 | m15 | y125
	if startDayElem.Exists() {
		str := strings.ToLower(startDayElem.String())
		if strings.HasPrefix(str, "w") { // Sunday = 0, ..., 6
			weekDay, err := strconv.Atoi(strings.TrimPrefix(str, "w"))
			if err != nil {
				return defaultValue, err
			}

			if weekDay < 0 || weekDay > 6 {
				return defaultValue, errors.New("$startDay out of range")
			}

			currentWeekDay := int(nextTime.Weekday())
			if weekDay > currentWeekDay {
				nextTime = nextTime.AddDate(0, 0, weekDay-currentWeekDay)
			} else if weekDay < currentWeekDay {
				nextTime = nextTime.AddDate(0, 0, 7+weekDay-currentWeekDay)
			}
		} else if strings.HasPrefix(str, "m") { // 1, 2, ..., 30
			monthDay, err := strconv.Atoi(strings.TrimPrefix(str, "m"))
			if err != nil {
				return defaultValue, err
			}

			if monthDay < 1 || monthDay > 30 {
				return defaultValue, errors.New("$startDay out of range")
			}

			currentMonthDay := nextTime.Day()
			if monthDay > currentMonthDay {
				nextTime = nextTime.AddDate(0, 0, monthDay-currentMonthDay)
			} else if monthDay < currentMonthDay {
				nextTime = nextTime.AddDate(0, 1, monthDay-currentMonthDay)
			}
		} else if strings.HasPrefix(str, "y") { // 1, 2, ..., 365
			yearDay, err := strconv.Atoi(strings.TrimPrefix(str, "y"))
			if err != nil {
				return defaultValue, err
			}

			if yearDay < 1 || yearDay > 365 {
				return defaultValue, errors.New("$startDay out of range")
			}

			currentYearDay := nextTime.YearDay()
			if yearDay > currentYearDay {
				nextTime = nextTime.AddDate(0, 0, yearDay-currentYearDay)
			} else if yearDay < currentYearDay {
				nextTime = nextTime.AddDate(1, 0, yearDay-currentYearDay)
			}
		}
	}
	startDayTimeElem := self.json.Get("$startDayTime") // 19:00:00
	if startDayTimeElem.Exists() {
		str := startDayTimeElem.String()
		timeOfDay, err := time.Parse("15:04:05", str)
		if err != nil {
			return defaultValue, errors.New("Parse $startDayTime error")
		}

		h, m, s := timeOfDay.Clock()
		dur := time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second
		ch, cm, cs := nextTime.Clock()
		currentDur := time.Duration(ch)*time.Hour + time.Duration(cm)*time.Minute + time.Duration(cs)*time.Second

		if dur < currentDur {
			nextTime = nextTime.AddDate(0, 0, 1)
		}
		nextTime = nextTime.Add(dur - currentDur)
	}

	// Calculate Delay, zero means job should be executed immediately
	return nextTime.Sub(now), nil
}

func (self *Rule) GetEntries() gjson.Result {
	return self.json.Get("$entries")
}

func (self *Rule) GetTargetTemplate(name string) gjson.Result {
	return self.json.Get(name)
}
