package elencho

import (
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"log"
)

type Department struct {
	Id   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type Degree struct {
	Id   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type StudyPlan struct {
	Id   string `json:"id"`
	Key  string `json:"key"`
	Year string `json:"year"`
}

const baseUrl = "https://www.unibz.it/en/timetable/PowerToolsForm/field"

func (db *Database) ClearTables() {
	err := db.Truncate([]string{
		"degree",
		"study_plan",
	})
	if err != nil {
		log.Fatalf("%q", err)
	}
}

func (db *Database) GetDepartments(departmentKey string) []Department {
	query := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("department")

	if departmentKey != noValue {
		query = query.Where(sq.Eq{"department_key": departmentKey})
	}

	rows, _ := db.Select(query, func(rows *sql.Rows) interface{} {
		var id, key, name string
		rows.Scan(&id, &key, &name)
		return Department{
			Id:   id,
			Key:  key,
			Name: name,
		}
	})

	departments := make([]Department, 0)
	for _, v := range rows {
		departments = append(departments, v.(Department))
	}

	return departments
}

func (db *Database) GetDegrees(departmentId string, degreeKey string) []Degree {
	query := sq.Select("*").From("degree")

	if departmentId != noValue {
		query = query.Where(sq.Eq{"department_fk": departmentId})
	}

	if degreeKey != noValue {
		query = query.Where(sq.Eq{"degree_key": degreeKey})
	}

	rows, _ := db.Select(query, func(rows *sql.Rows) interface{} {
		var id, fk, key, name string
		rows.Scan(&id, &fk, &key, &name)
		return Degree{
			Id:   id,
			Key:  key,
			Name: name,
		}
	})

	degrees := make([]Degree, 0)
	for _, v := range rows {
		degrees = append(degrees, v.(Degree))
	}

	return degrees
}

func (db *Database) GetStudyPlans(degreeId string, studyPlanKey string) []StudyPlan {
	query := sq.Select("*").From("study_plan")

	if degreeId != noValue {
		query = query.Where(sq.Eq{"degree_fk": degreeId})
	}

	if studyPlanKey != noValue {
		query = query.Where(sq.Eq{"study_plan_key": studyPlanKey})
	}

	rows, _ := db.Select(query, func(rows *sql.Rows) interface{} {
		var id, fk, key, year string
		rows.Scan(&id, &fk, &key, &year)
		return StudyPlan{
			Id:   id,
			Key:  key,
			Year: year,
		}
	})

	studyPlans := make([]StudyPlan, 0)
	for _, v := range rows {
		studyPlans = append(studyPlans, v.(StudyPlan))
	}

	return studyPlans
}

func (db *Database) InsertDegrees(department Department, degrees []Degree) {
	if len(degrees) > 0 {
		query := sq.Insert("degree").Columns("department_fk", "degree_key", "degree_name")

		for _, v := range degrees {
			query = query.Values(department.Id, v.Key, v.Name)
		}

		db.Insert(query)
	}
}

func (db *Database) InsertStudyPlans(degree Degree, studyPlans []StudyPlan) {
	if len(studyPlans) > 0 {
		query := sq.Insert("study_plan").Columns("degree_fk", "study_plan_key", "study_plan_year")

		for _, v := range studyPlans {
			query = query.Values(degree.Id, v.Key, v.Year)
		}

		db.Insert(query)
	}
}

func ParseAndInsertDegrees(db *Database, department Department) {
	degrees := make([]Degree, 0)

	for _, v := range connect(fmt.Sprintf("%s/degree/load?val=%s", baseUrl, department.Key)) {
		degrees = append(degrees, Degree{
			Id:   "",
			Key:  v["k"].(string),
			Name: v["v"].(string),
		})
	}

	db.InsertDegrees(department, degrees)
}

func ParseAndInsertStudyPlans(db *Database, degree Degree) {
	studyPlans := make([]StudyPlan, 0)

	for _, v := range connect(fmt.Sprintf("%s/studyPlan/load?val=%s", baseUrl, degree.Key)) {
		studyPlans = append(studyPlans, StudyPlan{
			Id:   "",
			Key:  v["k"].(string),
			Year: v["v"].(string),
		})
	}

	db.InsertStudyPlans(degree, studyPlans)
}
