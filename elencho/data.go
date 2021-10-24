package elencho

import (
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/gocolly/colly"
	_ "github.com/lib/pq"
	"log"
	"strconv"
	"strings"
	t "time"
)

// In both Urls we use English as language. For now we will support only
// English and further in the future new language support will be added.
const timetableFormBaseUrl = "https://www.unibz.it/en/timetable/PowerToolsForm/field"

// CSS queries used by the scraper to find specific course data in the website.
const allDaysQuery = "article"
const allCoursesQuery = ".u-pbi-avoid"
const dayDateQuery = "h2"
const courseRoomQuery = ".u-push-btm-quarter"
const courseDescriptionQuery = ".u-push-btm-1"
const courseProfessorQuery = ".actionLink"
const courseTimeAndType = ".u-push-btm-none:first-of-type"

// Time formats.
const inputDateTimeFormat = "Monday, 02 Jan 2006 15:04"
const outputDateTimeFormat = "2006-01-02 15:04"
const unibzDateFormat = "2006-01-02"

// Other constants.
const space = " "
const nothing = ""
const dot = "·"
const minus = "-"
const newLine = "\n"
const notAvailable = "N/A"

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

type Course struct {
	Start       JSONTime `json:"start"`
	End         JSONTime `json:"end"`
	Room        string `json:"room"`
	Description string `json:"description"`
	Professor   string `json:"professor"`
	Type        string `json:"type"`
}

type JSONTime struct {
	t.Time
}

func (t JSONTime)MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", t.Format(outputDateTimeFormat))
	return []byte(stamp), nil
}

func (db *Database) ClearTables() {
	err := db.Truncate([]string{
		"degree",
		"study_plan",
	})
	if err != nil {
		log.Fatalf("%q", err)
	}
}

func (db *Database) GetDepartments(departmentKey string) ([]Department, error) {
	query := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("department")

	if departmentKey != noValue {
		query = query.Where(sq.Eq{"department_key": departmentKey})
	}

	rows, err := db.Select(query, func(rows *sql.Rows) (interface{}, error) {
		var id, key, name string
		err := rows.Scan(&id, &key, &name)
		if err != nil {
			return nil, err
		}

		return Department{
			Id:   id,
			Key:  key,
			Name: name,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	departments := make([]Department, 0)
	for _, v := range rows {
		departments = append(departments, v.(Department))
	}

	return departments, nil
}

func (db *Database) GetDegrees(departmentId string, degreeKey string) ([]Degree, error) {
	query := sq.Select("*").From("degree")

	if departmentId != noValue {
		query = query.Where(sq.Eq{"department_fk": departmentId})
	}

	if degreeKey != noValue {
		query = query.Where(sq.Eq{"degree_key": degreeKey})
	}

	rows, err := db.Select(query, func(rows *sql.Rows) (interface{}, error) {
		var id, fk, key, name string
		err := rows.Scan(&id, &fk, &key, &name)
		if err != nil {
			return nil, err
		}

		return Degree{
			Id:   id,
			Key:  key,
			Name: name,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	degrees := make([]Degree, 0)
	for _, v := range rows {
		degrees = append(degrees, v.(Degree))
	}

	return degrees, nil
}

func (db *Database) GetStudyPlans(degreeId string, studyPlanKey string) ([]StudyPlan, error) {
	query := sq.Select("*").From("study_plan")

	if degreeId != noValue {
		query = query.Where(sq.Eq{"degree_fk": degreeId})
	}

	if studyPlanKey != noValue {
		query = query.Where(sq.Eq{"study_plan_key": studyPlanKey})
	}

	rows, err := db.Select(query, func(rows *sql.Rows) (interface{}, error) {
		var id, fk, key, year string
		err := rows.Scan(&id, &fk, &key, &year)
		if err != nil {
			return nil, err
		}

		return StudyPlan{
			Id:   id,
			Key:  key,
			Year: year,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	studyPlans := make([]StudyPlan, 0)
	for _, v := range rows {
		studyPlans = append(studyPlans, v.(StudyPlan))
	}

	return studyPlans, nil
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

	for _, v := range connect(fmt.Sprintf("%s/degree/load?val=%s", timetableFormBaseUrl, department.Key)) {
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

	for _, v := range connect(fmt.Sprintf("%s/studyPlan/load?val=%s", timetableFormBaseUrl, degree.Key)) {
		studyPlans = append(studyPlans, StudyPlan{
			Id:   "",
			Key:  v["k"].(string),
			Year: v["v"].(string),
		})
	}

	db.InsertStudyPlans(degree, studyPlans)
}

func GetDailyCourses(url string, deviceTime t.Time) ([]Course, error) {
	courses := make([]Course, 0)

	time := computeUnibzDateAsString(deviceTime)
	url = fmt.Sprintf("%s/?fromDate=%s&toDate=%s", url, time, time)
	fmt.Printf("scraping courses at %s", url)

	err := Scrape(url, allDaysQuery, func(e *colly.HTMLElement) {
		prevRoom := nothing
		day := e.ChildText(dayDateQuery)
		year := strconv.FormatInt(int64(t.Now().Year()), 10)

		e.ForEach(allCoursesQuery, func(i int, e *colly.HTMLElement) {
			course := Course{}

			courseStartTime, courseEndTime, courseType := getCourseTimeAndType(e)

			start, err := computeCourseDateTime(day, year, courseStartTime)
			if err == nil {
				course.Start = JSONTime{*start}
			}

			end, err := computeCourseDateTime(day, year, courseEndTime)
			if err == nil {
				course.End = JSONTime{*end}
			}

			courseRoom := e.ChildText(courseRoomQuery)
			if len(courseRoom) > 0 {
				course.Room = courseRoom
				prevRoom = courseRoom
			} else {
				course.Room = prevRoom
			}

			course.Description = e.ChildText(courseDescriptionQuery)
			course.Professor = e.ChildText(courseProfessorQuery)
			course.Type = courseType

			courses = append(courses, course)
		})
	})
	if err != nil {
		return nil, err
	}

	return courses, nil
}

func getCourseTimeAndType(e *colly.HTMLElement) (string, string, string) {
	startTime := notAvailable
	endTime  := notAvailable
	cType := notAvailable

	text := e.ChildText(courseTimeAndType)
	text = strings.ReplaceAll(text, space, nothing)
	text = strings.ReplaceAll(text, newLine, nothing)

	timesAndType := strings.Split(text, dot)
	if len(timesAndType) > 1 {
		courseTimes := strings.Split(timesAndType[0], minus)
		if len(courseTimes) > 1 {
			startTime = courseTimes[0]
			endTime = courseTimes[1]
		}

		cType = timesAndType[1]
	}

	return startTime, endTime, cType
}
