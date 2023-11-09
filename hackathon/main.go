package main

import (
	"encoding/json"
	"net/http"
)

type responseMessage struct {
	Message string `json:"message"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	v := r.URL.Query().Get("name")
	if v == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	bytes, err := json.Marshal(responseMessage{
		Message: "Hello," + v + "!",
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

func main() {
	http.HandleFunc("/hello", handler)
	http.ListenAndServe(":8080", nil)
}
