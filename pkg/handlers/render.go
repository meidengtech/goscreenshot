package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/sempr/goscreenshot/constants"
	"github.com/sempr/goscreenshot/pkg/lru"
	"github.com/sempr/goscreenshot/pkg/shot"
)

// RenderArgs xxx
type RenderArgs struct {
	Width int    `schema:"width"`
	HTML  string `schema:"html"`
}

// ShotApp holding the request and the shot object
type ShotApp struct {
	Shot *shot.PooledShotter
}

func randInt() int {
	now := time.Now()
	todaySecend := now.Unix() % 3600
	return int(todaySecend)*1000000 + rand.Intn(1000000)
}

// AddRoutes Add Routes
func (app *ShotApp) AddRoutes(r *mux.Router) {
	r.HandleFunc("/render", app.RenderHandler)
}

// RenderHandler 是图片渲染的主入口
func (app *ShotApp) RenderHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var args = new(RenderArgs)
	decoder := schema.NewDecoder()
	err = decoder.Decode(args, r.Form)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//t, err := template.ParseFiles("../template/model.html")

	wrappedHTMLBase := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf8" />
<style>
html, body, div, span, applet, object, iframe,
h1, h2, h3, h4, h5, h6, p, blockquote, pre,
a, abbr, acronym, address, big, cite, code,
del, dfn, em, img, ins, kbd, q, s, samp,
small, strike, strong, sub, sup, tt, var,
b, u, i, center,
dl, dt, dd, ol, ul, li,
fieldset, form, label, legend,
table, caption, tbody, tfoot, thead, tr, th, td,
article, aside, canvas, details, embed,
figure, figcaption, footer, header, hgroup,
menu, nav, output, ruby, section, summary,
time, mark, audio, video {
        margin: 0;
        padding: 0;
        border: 0;
        font-size: 100%;
        font: inherit;
        vertical-align: baseline;
}
/* HTML5 display-role reset for older browsers */
article, aside, details, figcaption, figure,
footer, header, hgroup, menu, nav, section {
        display: block;
}
body {
        line-height: 1;
}
ol, ul {
        list-style: none;
}
blockquote, q {
        quotes: none;
}
blockquote:before, blockquote:after,
q:before, q:after {
        content: '';
        content: none;
}
table {
        border-collapse: collapse;
        border-spacing: 0;
}
</style>

</head>
<body>
<script>
    document.addEventListener("DOMContentLoaded", function() {
        document.removeEventListener("DOMContentLoaded", arguments.callee, false);
        var imgs = document.getElementsByTagName("img");
        var f = function() {
            var complete = true;
            for (var i = 0; i != imgs.length; i ++) {
                if (!imgs[i].complete) {
                    complete = false;
                    break;
                }
            }
            if (complete) {
                document.getElementById('ImgLoadedFlagACHHcLIkD3').style.display = "block";
            } else {
                window.setTimeout(f, 50);
            }
        };
        f();
        window.setTimeout(function() {
        	document.getElementById('ImgLoadedFlagACHHcLIkD3').style.display = "block";
        }, 3000);
    });
</script>
<div id="ImgLoadedFlagACHHcLIkD3" style="display:none;">test</div>
<div id="ACHHcLIkD3">
`

	wrappedHTML := wrappedHTMLBase + args.HTML
	key := fmt.Sprintf("%d", randInt())
	lru.SavedHTMLMap.Add(key, wrappedHTML)
	pageURL := fmt.Sprintf("http://127.0.0.1:%d/html/%s", constants.ServerPort, key)
	if args.Width > 2000 || args.Width < 10 {
		args.Width = 750
	}
	picbuf, err := app.Shot.Screenshot(pageURL, args.Width)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "image/png")
	w.Header().Set("content-length", fmt.Sprintf("%d", len(picbuf)))
	w.Write(picbuf)
}
