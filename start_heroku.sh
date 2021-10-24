#!/bin/bash
echo "Which build of heroku? [local, local web, local worker]"
read local dyno
go build -o bin/elencho-scraper-web -v cmd/elencho-scraper-web/main.go
go build -o bin/elencho-scraper-worker -v cmd/elencho-scraper-worker/main.go
heroku $local $dyno
