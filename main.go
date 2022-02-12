package main

import (
	"fmt"
	"log"
	"net/http"
)

func hello(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprint(w, "hello world")
	if err != nil {
		log.Fatalf("Error in hello world: %s", err)
		return
	}
}
func main() {
	http.HandleFunc("/hello", hello)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatalf("Error in setting up server: %s", err)
		return
	}
}
