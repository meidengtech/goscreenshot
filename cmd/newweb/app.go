package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	lru "github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
)

type app struct {
	shot IShot
	lru  *lru.Cache
	log  *logrus.Logger
	port int
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
	a.log.Infof("Received: html %s width %d.", args.HTML, args.Width)
	key := fmt.Sprintf("%d", rand.Intn(1000000))

	a.lru.Add(key, args.HTML)
	pageURL := fmt.Sprintf("http://127.0.0.1:%d/html/%s", a.port, key)
	fmt.Println(pageURL)
	if args.Width > 2000 || args.Width < 10 {
		args.Width = 750
	}
	a.log.Debug("start screenshot")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
	defer cancel()

	picbuf, err := a.shot.Do(ctx, pageURL, args.HTML, args.Width)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
        a.log.Infof("Failed: html %s width %d.", args.HTML, args.Width)
		a.log.Error(err)
		fmt.Fprintln(w, err)
		return
	}
	a.log.Infof("Succeed: html %s width %d.", args.HTML, args.Width)
	w.Header().Set("content-type", "image/jpeg")
	w.Header().Set("content-length", fmt.Sprintf("%d", len(picbuf)))
	w.Write(picbuf)
}

func (a *app) Stat(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(a.shot.Stat()))
}
