package elencho

import (
	"fmt"
	t "time"
)

func computeCourseDateTime(day string, year string, time string) (*t.Time, error) {
	return convertStringToTime(fmt.Sprintf("%s %s %s", day, year, time), inputDateTimeFormat)
}

func computeDeviceTime(deviceTime string) (*t.Time, error) {
	return convertStringToTime(deviceTime, outputDateTimeFormat)
}

func convertStringToTime(date string, format string) (*t.Time, error) {
	result, err := t.Parse(format, date)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func computeUnibzDateAsString(time t.Time) string {
	return convertTimeToString(time, unibzDateFormat)
}

func convertTimeToString(time t.Time, format string) string {
	return time.Format(format)
}
