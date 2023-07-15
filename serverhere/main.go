package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

func main() {
	mux := chi.NewRouter()
	mux.Use(middle)
	mux.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir(".")).ServeHTTP(w, r)
	})
	server := http.Server{
		Addr:    ":8888",
		Handler: mux,
	}

	log.Printf("Starting Server on port %q\n", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}

func middle(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%v\n", r.URL.RequestURI())
			next.ServeHTTP(w, r)
		},
	)
}
