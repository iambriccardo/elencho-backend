package elencho

import (
	"fmt"
	"os"
	"strconv"
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

func GetEnv(key string) (string, error) {
	variable := os.Getenv(key)
	if variable == "" {
		return "", fmt.Errorf("variable %s is not found in the .env file", variable)
	}

	return variable, nil
}

func GetIntEnv(key string) (int, error) {
	variable, err := GetEnv(key)
	if err != nil {
		return 0, err
	}

	variableInt, err := strconv.Atoi(variable)
	if err != nil {
		return 0, err
	}

	return variableInt, err
}

func DefaultGetIntEnv(key string, defaultValue int) int {
	variable, err := GetIntEnv(key)
	if err != nil {
		variable = defaultValue
	}

	return variable
}
