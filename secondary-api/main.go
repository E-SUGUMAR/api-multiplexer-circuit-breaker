package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Response from SECONDARY API")
}

func main() {
	http.HandleFunc("/", handler)

	fmt.Println("Secondary API running on :8082")

	log.Fatal(http.ListenAndServe(":8082", nil))
}
