package main

import (
	"log"
	"os"

	"github.com/RiccardoBusetti/elencho-scraper/elencho"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

type Endpoint struct {
	RelativePath string
	Handler      func(*gin.Context)
}

func getRoutes(db *elencho.Database) []Endpoint {
	return []Endpoint{
		{
			RelativePath: "/",
			Handler: func(c *gin.Context) {
				c.JSON(200, gin.H{
					"Status": "The service is up and running correctly.",
				})
			},
		},
		{
			RelativePath: "/departments",
			Handler: func(c *gin.Context) {
				d, err := elencho.Departments(db)
				if err != nil {
					c.JSON(500, err)
				} else {
					c.JSON(200, d)
				}
			},
		},
		{
			RelativePath: "/degrees",
			Handler: func(c *gin.Context) {
				departmentId := c.DefaultQuery("departmentId", "")
				d, err := elencho.Degrees(db, departmentId)
				if err != nil {
					c.JSON(500, err)
				} else {
					c.JSON(200, d)
				}
			},
		},
		{
			RelativePath: "/studyPlans",
			Handler: func(c *gin.Context) {
				degreeId := c.DefaultQuery("degreeId", "")
				s, err := elencho.StudyPlans(db, degreeId)
				if err != nil {
					c.JSON(500, err)
				} else {
					c.JSON(200, s)
				}
			},
		},
	}
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatalf("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())

	db := elencho.Make()
	db.Open()
	defer db.Close()

	for _, e := range getRoutes(db) {
		router.GET(e.RelativePath, e.Handler)
	}

	router.Run(":" + port)
}
