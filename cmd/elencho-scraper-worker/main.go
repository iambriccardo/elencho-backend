package main

import (
	"fmt"
	"github.com/RiccardoBusetti/elencho-scraper/elencho"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const oneWeekHours = 168

func main() {
	tick := time.NewTicker(time.Hour * oneWeekHours)
	done := make(chan bool)
	go scheduler(tick, done)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-sigs
	done <- true
}

func scheduler(tick *time.Ticker, done chan bool) {
	for {
		select {
		case <-tick.C:
			task()
		case <-done:
			return
		}
	}
}

func task() {
	db := elencho.Make()
	err := db.Open()
	if err != nil {
		fmt.Printf("an error occurred while running the scraper: %q", err)
	}
	defer db.Close()
	err = elencho.Start(db)
	if err != nil {
		fmt.Printf("an error occurred while running the scraper: %q", err)
	}
}
