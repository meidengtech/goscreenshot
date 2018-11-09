package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

var (
	nWorkers = flag.Int("n", 4, "The number of workers to start")
	port     = flag.Int("p", 8090, "The port of the http server")
)

func handleSignal(s chan os.Signal, shot *QueuedShotter) {
	sig := <-s
	shot.Stop()
	fmt.Println(sig)
	time.Sleep(3)
	os.Exit(1)
}

func getLog() *logrus.Logger {
	customFormatter := logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	}
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log := logrus.New()
	log.Formatter = &customFormatter
	return log
}

func getApp() *negroni.Negroni {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	signal.Notify(ch, syscall.SIGTERM)

	log := getLog()
	lruSize := 4096
	l, err := lru.New(lruSize)
	if err != nil {
		log.Fatal(err)
	}
	shot := QueuedShotter{workNum: *nWorkers, debugServer: "http://127.0.0.1:9222"}
	shot.Start()

	go handleSignal(ch, &shot)
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
	logrus.Fatal(http.ListenAndServe(":8090", neg))
}
