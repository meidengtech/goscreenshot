package lru

import (
	"log"

	"github.com/hashicorp/golang-lru"
)

// SavedHTMLMap save recent html codes
var SavedHTMLMap *lru.Cache

func init() {
	l, err := lru.New(256)
	if err != nil {
		log.Fatal(err)
	}
	SavedHTMLMap = l
}
