package elencho

import (
	"fmt"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"log"
	"sort"
	"time"
)

const databaseUrlEnv = "DATABASE_URL"
const noValue = ""
const timetableBaseUrl = "https://www.unibz.it/en/timetable"

func Start(db *Database) error {
	log.Printf("starting preparing the courses database")
	db.ClearTables()

	departments, err := db.GetDepartments("")
	if err != nil {
		return err
	}
	for _, department := range departments {
		ParseAndInsertDegrees(db, department)

		degrees, err := db.GetDegrees(department.Id, "")
		if err != nil {
			return err
		}

		for _, degree := range degrees {
			ParseAndInsertStudyPlans(db, degree)
		}
	}
	log.Println("finished preparing the courses database")
	return nil
}

func Departments(db *Database) ([]Department, error) {
	return db.GetDepartments("")
}

func Degrees(db *Database, departmentId string) ([]Degree, error) {
	return db.GetDegrees(departmentId, "")
}

func StudyPlans(db *Database, degreeId string) ([]StudyPlan, error) {
	return db.GetStudyPlans(degreeId, "")
}

func CheckAvailability(room string, deviceTime string) (map[string]interface{}, error) {
	if room == noValue || deviceTime == noValue {
		return nil, fmt.Errorf("error while checking availability: you must choose a room and your current time")
	}

	log.Printf("checking availability for room %s from time %s", room, deviceTime)
	courses, err := GetCourses(timetableBaseUrl)
	if err != nil {
		return nil, fmt.Errorf("error while checking availability: %q", err)
	}

	rooms := getRooms(courses)
	matches := fuzzy.RankFind(room, rooms)
	sort.Sort(matches)
	if len(matches) > 0 {
		log.Printf("estimation of room %s is %s", room, matches[0].Target)
		room = matches[0].Target
	}

	deviceTimeConverted, err := computeDeviceTime(deviceTime)
	if err != nil {
		return nil, fmt.Errorf("error while checking availability: %q", err)
	}

	courses = getCoursesByRoom(courses, room, *deviceTimeConverted)

	log.Printf("computing available time slots")
	return map[string]interface{}{
		"room":           room,
		"availabilities": getAvailableTimeSlots(courses, *deviceTimeConverted),
	}, nil
}

func getRooms(courses []Course) []string {
	rooms := make([]string, 0)

	for _, v := range courses {
		rooms = append(rooms, v.Room)
	}

	return rooms
}

func getCoursesByRoom(courses []Course, roomName string, deviceTime time.Time) []Course {
	fCourses := make([]Course, 0)
	for _, v := range courses {
		if v.Room == roomName {
			fCourses = append(fCourses, v)
		}
	}

	return fCourses
}

func getAvailableTimeSlots(courses []Course, deviceTime time.Time) []map[string]interface{} {
	availableTimeSlots := make([]map[string]interface{}, 0)

	courses = computeBusyTimeSlots(courses)

	if len(courses) > 0 {
		availableTimeSlots = append(availableTimeSlots, map[string]interface{}{
			"from": nil,
			"to":   courses[0].Start,
		})

		for i := 0; i < len(courses)-1; i++ {
			course1 := courses[i]
			course2 := courses[i+1]

			if !haveSameTime(course1, course2) && havePause(course1, course2) {
				availableTimeSlots = append(availableTimeSlots, map[string]interface{}{
					"from": course1.End,
					"to":   course2.Start,
				})
			}
		}

		availableTimeSlots = append(availableTimeSlots, map[string]interface{}{
			"from": courses[len(courses)-1].End,
			"to":   nil,
		})
	}

	return availableTimeSlots
}

func computeBusyTimeSlots(courses []Course) []Course {
	filteredCourses := make([]Course, 0)

	for _, course1 := range courses {
		if len(filteredCourses) == 0 {
			filteredCourses = append(filteredCourses, course1)
		} else {
			found := false
			i := 0
			for i < len(filteredCourses) && !found {
				course2 := filteredCourses[i]
				if haveSameTime(course1, course2) {
					found = true
				} else if isWithinOtherCourse(course1, course2) {
					found = true
				} else if isLongerThanOtherCourse(course1, course2) {
					filteredCourses[i] = course1
					found = true
				} else if isOverlappingWithOtherCourse(course1, course2) {
					filteredCourses[i] = Course{
						Start: course2.Start,
						End:   course1.End,
					}
					found = true
				}
				i++
			}

			if !found {
				filteredCourses = append(filteredCourses, course1)
			}
		}
	}

	return filteredCourses
}

func haveSameTime(course1 Course, course2 Course) bool {
	return course1.Start.Equal(course2.Start) && course1.End.Equal(course2.End)
}

func isWithinOtherCourse(course1 Course, course2 Course) bool {
	return course1.Start.Equal(course2.Start) || course1.Start.After(course2.Start) && course1.End.Before(course2.End)
}

func isLongerThanOtherCourse(course1 Course, course2 Course) bool {
	return course1.Start.Equal(course2.Start) && course1.End.After(course2.End)
}

func isOverlappingWithOtherCourse(course1 Course, course2 Course) bool {
	return course1.Start.After(course2.Start) && course1.End.Before(course2.Start) && course1.End.After(course2.End)
}

func havePause(course1 Course, course2 Course) bool {
	return !course1.End.Equal(course2.Start)
}

func isCourseFinished(course Course, deviceTime time.Time) bool {
	return deviceTime.After(course.End)
}

func isCourseNow(course Course, deviceTime time.Time) bool {
	return (deviceTime.Equal(course.Start) || deviceTime.After(course.Start)) &&
		(deviceTime.Equal(course.End) || deviceTime.Before(course.End))
}

func isCourseUpcoming(course Course, deviceTime time.Time) bool {
	return deviceTime.Before(course.Start)
}
