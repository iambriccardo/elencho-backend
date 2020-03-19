package elencho

import (
	"fmt"
	t "time"
)

func computeCourseDateTime(day string, year string, time string) (*t.Time, error) {
	result, err := t.Parse(inputDateTimeFormat, fmt.Sprintf("%s %s %s", day, year, time))
	if err != nil {
		return nil, err
	}
	return &result, nil
}
