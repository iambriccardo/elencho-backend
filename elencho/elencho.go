package elencho

const databaseUrlEnv = "DATABASE_URL"
const noValue = ""

func Start(db *Database) {
	db.ClearTables()
	for _, department := range db.GetDepartments("") {
		ParseAndInsertDegrees(db, department)
		for _, degree := range db.GetDegrees(department.Id, "") {
			ParseAndInsertStudyPlans(db, degree)
		}
	}
}
