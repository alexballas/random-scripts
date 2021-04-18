package main

import (
	"log"
	"net/http"
)

func main() {
	server := http.Server{
		Addr: ":8888",
	}
	mux := http.NewServeMux()
	server.Handler = middle(mux)
	mux.Handle("/", http.FileServer(http.Dir(".")))
	log.Printf("Starting Server on port %q\n", server.Addr)
	server.ListenAndServe()
}

func middle(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%v\n", r.URL.RequestURI())
			next.ServeHTTP(w, r)
		},
	)
}
