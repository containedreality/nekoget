package server

import (
	"log"
	"net/http"
)

func filehandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.RemoteAddr, r.URL)

		handler.ServeHTTP(w, r)
	})
}

func ServeFiles(addr string, path string) error {
	mux := http.NewServeMux()
	filesystem := http.FileServer(http.Dir(path))
	mux.Handle("GET /", filehandler(filesystem))

	return http.ListenAndServe(addr, mux)
}
