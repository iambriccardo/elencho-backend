package elencho

import (
	"log"
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

func CheckAvailability() ([]Course, error) {
	return GetCourses(timetableBaseUrl)
}
