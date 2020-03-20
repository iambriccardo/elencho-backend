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
		if v.Room == roomName && (isCourseNow(v, deviceTime) || isCourseUpcoming(v, deviceTime)) {
			fCourses = append(fCourses, v)
		}
	}

	return fCourses
}

func getAvailableTimeSlots(courses []Course, deviceTime time.Time) []map[string]interface{} {
	availableTimeSlots := make([]map[string]interface{}, 0)

	if len(courses) > 0 {
		if isCourseUpcoming(courses[0], deviceTime) {
			availableTimeSlots = append(availableTimeSlots, map[string]interface{}{
				"from": "",
				"to":   courses[0].Start,
			})
		}

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
			"to":   noValue,
		})
	}

	return availableTimeSlots
}

// TODO: check also if one course is withing the time of the other and vice versa.
func haveSameTime(course1 Course, course2 Course) bool {
	return course1.Start.Equal(course2.Start) && course1.End.Equal(course2.End)
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
