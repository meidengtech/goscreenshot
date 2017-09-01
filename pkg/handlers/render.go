package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
	//"html/template"

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

func randInt() int {
	now := time.Now()
	todaySecend := now.Unix() % 3600
	return int(todaySecend)*1000 + rand.Intn(1000)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

// RenderHandler 是图片渲染的主入口
func RenderHandler(w http.ResponseWriter, r *http.Request) {
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
%s
</style>
</head>
<body>
<div id="ACHHcLIkD3">
%s
</div>
<div id="ImgLoadedFlagACHHcLIkD3" style="display:none;">test</div>
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
                window.setTimeout(f);
            }
        };
        f();
    });
</script>
</body>
</html>`
	wrappedHTML := fmt.Sprintf(wrappedHTMLBase, constants.RESETCSS, args.HTML)
	key := fmt.Sprintf("%d", randInt())
	lru.SavedHTMLMap.Add(key, wrappedHTML)
	pageURL := fmt.Sprintf("http://127.0.0.1:%d/html/%s", constants.ServerPort, key)
	picbuf, err := shot.Screenshot(pageURL, args.Width)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "image/png")
	w.Header().Set("content-length", fmt.Sprintf("%d", len(picbuf)))
	w.Write(picbuf)
}
