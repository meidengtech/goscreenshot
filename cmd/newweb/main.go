package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gorilla/mux"
	"github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

var (
	NWorkers = flag.Int("n", 4, "The number of workers to start")
	HTTPAddr = flag.String("http", "127.0.0.1:8000", "Address to listen for HTTP requests on")
)

func handleSignal(s chan os.Signal, shot *QueuedShotter) {
	sig := <-s
	shot.Stop()
	fmt.Println(sig)
	time.Sleep(3)
	os.Exit(1)
}

func getApp() *negroni.Negroni {

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	signal.Notify(ch, syscall.SIGTERM)

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
	shot := QueuedShotter{cPool: p, workNum: *NWorkers}
	shot.Start()
	go handleSignal(ch, &shot)
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
	flag.Parse()
	neg := getApp()
	log.Fatal(http.ListenAndServe(":8090", neg))
}
