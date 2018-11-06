package main

import (
	"log"
	"net/http"

	"github.com/chromedp/chromedp"
	"github.com/gorilla/mux"
	"github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

type favContextKey string

func getApp() *negroni.Negroni {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log := logrus.New()
	log.Formatter = customFormatter
	// log.SetLevel(logrus.InfoLevel)
	// log.SetLevel(logrus.DebugLevel)
	// port := 8090
	// portStr := fmt.Sprintf(":%d", port)
	// new lru cache
	lruSize := 4096

	l, err := lru.New(lruSize)
	if err != nil {
		log.Fatal(err)
	}
	// new chromedp pool
	chromePortStart, chromePortEnd := 50000, 51000
	p, err := chromedp.NewPool(
		chromedp.PoolLog(log.Infof, log.Debugf, log.Errorf),
		chromedp.PortRange(chromePortStart, chromePortEnd),
	)
	// object pool
	// shot := PooledShotter2{cPool: p, log: log}
	shot := QueuedShotter{cPool: p}
	shot.Start()
	if err != nil {
		log.Fatal(err)
	}

	a := app{shot: &shot, lru: l, log: log}

	r := mux.NewRouter()
	r.HandleFunc("/html/{id:[0-9]+}", a.ContentPage)
	r.HandleFunc("/", a.IndexPage)
	r.HandleFunc("/render", a.Render)

	neg := negroni.Classic()
	neg.UseHandler(r)
	return neg
}

func main() {
	neg := getApp()
	log.Fatal(http.ListenAndServe(":8090", neg))
}
