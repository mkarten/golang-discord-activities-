package main

import (
	"log"
	"net/http"
)

func serverFolderNoCache(dir http.Dir) http.Handler {
	return http.StripPrefix("/", http.FileServer(http.Dir(dir)))
}

func main() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", noCache(fs))

	log.Print("Listening on :3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
		w.Header().Set("Expires", "0")                                         // Proxies.
		h.ServeHTTP(w, r)
	})
}
