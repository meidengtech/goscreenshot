package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	lru "github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
)

type app struct {
	shot IShot
	lru  *lru.Cache
	log  *logrus.Logger
}

func (a *app) IndexPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Index"))
}

func (a *app) ContentPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	val, has := a.lru.Get(vars["id"])
	if !has {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Write([]byte(wrappedHTMLBase + val.(string)))
}

type renderArgs struct {
	Width int    `schema:"width"`
	HTML  string `schema:"html"`
}

func (a *app) Render(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		a.log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}
	var args = new(renderArgs)
	decoder := schema.NewDecoder()
	err = decoder.Decode(args, r.Form)
	if err != nil {
		a.log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}
	a.log.Debugln(args.HTML, args.Width)
	key := fmt.Sprintf("%d", rand.Intn(1000000))

	a.lru.Add(key, args.HTML)
	pageURL := fmt.Sprintf("http://127.0.0.1:8090/html/%s", key)
	fmt.Println(pageURL)
	if args.Width > 2000 || args.Width < 10 {
		args.Width = 750
	}
	a.log.Debug("start screenshot")
	picbuf, err := a.shot.Do(context.TODO(), pageURL, args.Width)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}
	w.Header().Set("content-type", "image/jpeg")
	w.Header().Set("content-length", fmt.Sprintf("%d", len(picbuf)))
	w.Write(picbuf)
}
