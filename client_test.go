package main

import (
	"fmt"
	"net/http"
)

// SearchServer
func SearchServer(w http.ResponseWriter, r *http.Request) {
	myParam := r.URL.Query().Get("param")
	if myParam != "" {
		fmt.Fprintln(w, "`myParam` is", myParam)
	}

	key := r.FormValue("key")
	if key != "" {
		fmt.Fprintln(w, "`key` is", key)
	}
}

