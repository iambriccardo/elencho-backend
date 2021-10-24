package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	el "github.com/RiccardoBusetti/elencho-scraper/elencho"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	port, err := el.GetEnv("PORT")
	if err != nil {
		log.Fatalf("an error occurred in web: %q", err)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(CORSMiddleware())

	poolSize, err := el.GetIntEnv("POOL_SIZE")
	if err != nil {
		log.Fatalf("an error occurred in web: %q", err)
	}

	pool := make(chan func(r *el.Request, db *el.Database), poolSize)
	openPool(poolSize, pool)

	db := el.Make()
	err = db.Open()
	if err != nil {
		log.Fatalf("an error occurred in web: %q", err)
	}
	defer db.Close()

	var wg sync.WaitGroup
	for _, e := range el.EnabledEndpoints() {
		func(e el.EndPoint) {
			router.GET(e.String(), func(ctx *gin.Context) {
				select {
				case f := <-pool:
					wg.Add(1)
					go func() {
						f(&el.Request{EndPoint: e, Context: ctx}, db)
						wg.Done()
					}()
					wg.Wait()
					pool <- f
					break
				default:
					el.Response{
						Context: ctx,
						Error:   fmt.Errorf("the server rejected the request, because it is under heavy load"),
					}.WithError()
					break
				}
			})
		}(e)
	}

	err = router.Run(":" + port)
	if err != nil {
		close(pool)
		log.Fatalf("an error occurred in web: %q", err)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func openPool(poolSize int, pool chan<- func(r *el.Request, db *el.Database)) {
	for i := 0; i < poolSize; i++ {
		pool <- func(r *el.Request, db *el.Database) {
			responseChan := make(chan el.Response, 1)
			go handleRequest(r, db, responseChan)

			secondsTimeout := el.DefaultGetIntEnv("REQUEST_TIMEOUT_SECONDS", 0)

			ctx := r.Context

			select {
			case response := <-responseChan:
				if response.Error != nil {
					response.WithError()
				} else if response.Content != nil {
					response.WithSuccess()
				}
				break
			case <-time.After(time.Duration(secondsTimeout) * time.Second):
				el.Response{Context: ctx}.WithTimeout()
				break
			}
		}
	}
}

func handleRequest(r *el.Request, db *el.Database, responseChan chan<- el.Response) {
	baseResponse := el.Response{
		Context: r.Context,
	}

	switch r.EndPoint {
	case el.Base:
		baseResponse.Content = gin.H{"status": "The service is up and running."}
		break
	case el.GetDepartments:
		ds, err := el.Departments(db)
		if err != nil {
			baseResponse.Error = err
		} else {
			baseResponse.Content = ds
		}
		break
	case el.GetDegrees:
		departmentId := r.Context.DefaultQuery("departmentId", "")
		ds, err := el.Degrees(db, departmentId)
		if err != nil {
			baseResponse.Error = err
		} else {
			baseResponse.Content = ds
		}
		break
	case el.GetStudyPlans:
		degreeId := r.Context.DefaultQuery("degreeId", "")
		ss, err := el.StudyPlans(db, degreeId)
		if err != nil {
			baseResponse.Error = err
		} else {
			baseResponse.Content = ss
		}
		break
	case el.CheckAvailability:
		room := r.Context.DefaultQuery("room", "")
		deviceTime := r.Context.DefaultQuery("deviceTime", "")
		at, err := el.CheckRoomAvailability(room, deviceTime)
		if err != nil {
			baseResponse.Error = err
		} else {
			baseResponse.Content = at
		}
		break
	case el.Refresh:
		go el.Start(db) // We launch the work asynchronously.
		baseResponse.Content = gin.H{"status": "The refresh has been started."}
		break
	default:
		// We do not perform anything because the server is unable to handle such kind of request,
		// thus the timeout will trigger an error response.
		break
	}

	responseChan <- baseResponse
}
