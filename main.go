package main

import (
	"github.com/getsentry/sentry-go"
	"os"
	"strconv"
	"time"
)

func main() {
	host := os.Getenv("RANCHER_HOST")
	token := os.Getenv("RANCHER_TOKEN")
	project := os.Getenv("RANCHER_PROJECT")
	filepath := os.Getenv("FILE")
	period := 60 * time.Second
	if sec, err := strconv.Atoi(os.Getenv("PERIOD_SEC")); err == nil {
		period = time.Duration(sec) * time.Second
	}
	initSentry(os.Getenv("SENTRY_DSN"))

	rancher := NewRancher(host, token, project)
	writer := NewTargetWriter(filepath)

	for {
		targets, err := rancher.ListAutoPromServices()
		if err != nil {
			os.Stderr.WriteString("[ERROR] list rancher error: " + err.Error())
			sentry.CaptureException(err)
			time.Sleep(period)
			continue
		}
		err = writer.Write(targets)
		if err != nil {
			os.Stderr.WriteString("[ERROR] write file error: " + err.Error())
			sentry.CaptureException(err)
		}
		time.Sleep(period)
	}
}

func initSentry(dsn string) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:   dsn,
		Debug: false,
	})
	if err != nil {
		os.Stderr.WriteString("[ERROR] init sentry error: " + err.Error())
	}
}
