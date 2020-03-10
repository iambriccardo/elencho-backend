package elencho

import "log"

const databaseUrlEnv = "DATABASE_URL"
const noValue = ""

func Start(db *Database) {
	log.Printf("starting preparing the courses database")
	db.ClearTables()
	for _, department := range db.GetDepartments("") {
		ParseAndInsertDegrees(db, department)
		for _, degree := range db.GetDegrees(department.Id, "") {
			ParseAndInsertStudyPlans(db, degree)
		}
	}
	log.Println("finished preparing the courses database")
}
