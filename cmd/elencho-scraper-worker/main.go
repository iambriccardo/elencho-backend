package main

import (
	"github.com/RiccardoBusetti/elencho-scraper/elencho"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	tick := time.NewTicker(time.Hour * 24)
	done := make(chan bool)
	go scheduler(tick, done)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	done <- true
}

func scheduler(tick *time.Ticker, done chan bool) {
	task()
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
	db.Open()
	defer db.Close()
	elencho.Start(db)
}