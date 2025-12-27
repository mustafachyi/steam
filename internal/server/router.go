package server

import (
	"net/http"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/lookup", HandleLookup)
	mux.HandleFunc("/search", HandleSearch)
	return mux
}
