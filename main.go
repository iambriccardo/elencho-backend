package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/heroku/go-getting-started/elencho"
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
				elencho.Start(db)
				c.JSON(200, gin.H{
					"Status": "The service has successfully updated its database.",
				})
			},
		},
		{
			RelativePath: "/departments/:key",
			Handler: func(c *gin.Context) {

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
