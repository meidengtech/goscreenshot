package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sempr/goscreenshot/pkg/lru"
)

// PageHandler 提供一个链接用于Chrome进行截图操作
func PageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	val, has := lru.SavedHTMLMap.Get(vars["id"])
	if !has {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Write([]byte(val.(string)))
}
