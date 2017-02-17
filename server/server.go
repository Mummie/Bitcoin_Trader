package server

import (
	"io"
	"net/http"
)

func write(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello!")
}

func RunServer() {
	http.HandleFunc("/", write)
	http.ListenAndServe(":5000", nil)
}
