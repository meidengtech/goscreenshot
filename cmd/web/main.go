package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sempr/goscreenshot/pkg/handlers"
	"github.com/urfave/negroni"
)

func prepareWeb() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/html/{id:[0-9]+}", handlers.PageHandler)
	r.HandleFunc("/render", handlers.RenderHandler)
	return r
}

func main() {
	mux := prepareWeb()
	neg := negroni.Classic()
	neg.UseHandler(mux)
	log.Println("Starting Web Server on port 8019")
	log.Fatal(http.ListenAndServe(":8019", neg))
}
