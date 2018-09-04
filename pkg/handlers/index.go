package handlers

import "net/http"

// IndexHandler 是默认首页，无内容
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}
