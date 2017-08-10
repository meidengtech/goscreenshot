package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
	"github.com/sempr/goscreenshot/pkg/handlers"
	"github.com/sempr/goscreenshot/pkg/shot"
	"github.com/urfave/negroni"
)

func prepareWeb() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/html/{id:[0-9]+}", handlers.PageHandler)
	r.HandleFunc("/render", handlers.RenderHandler)
	return r
}

func handleSignal(s chan os.Signal) {
	sig := <-s
	fmt.Println(sig)
	shot.Release()
	os.Exit(1)
}

func main() {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)

	go handleSignal(ch)
	shot.Init()

	mux := prepareWeb()
	neg := negroni.Classic()
	neg.UseHandler(mux)
	log.Println("Starting Web Server on port 8019")
	log.Fatal(http.ListenAndServe(":8019", neg))
}
