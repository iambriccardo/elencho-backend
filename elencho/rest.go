package elencho

import "github.com/gin-gonic/gin"

type EndPoint int

const (
	Base = iota
	GetDepartments
	GetDegrees
	GetStudyPlans
	CheckAvailability
	Refresh
)

func EnabledEndpoints() []EndPoint {
	return []EndPoint{
		Base,
		GetDepartments,
		GetDegrees,
		GetStudyPlans,
		CheckAvailability,
		Refresh,
	}
}

func (e EndPoint) String() string {
	return [...]string{"/", "/departments", "/degrees", "/studyPlans", "/availability", "/refresh"}[e]
}

type Request struct {
	EndPoint EndPoint
	Context  *gin.Context
}

type Response struct {
	Content interface{}
	Context *gin.Context
	Error   error
}

func (r Response) WithSuccess() {
	r.Context.JSON(200, r.Content)
}

func (r Response) WithError() {
	r.Context.JSON(500, r.Error.Error())
	r.Context.Abort()
}

func (r Response) WithTimeout() {
	r.Context.JSON(504, "timeout")
	r.Context.Abort()
}
