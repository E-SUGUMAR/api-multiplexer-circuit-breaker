package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(100 * time.Millisecond)

	fmt.Fprintf(w, "Response from PRIMARY API")
}

func main() {
	http.HandleFunc("/", handler)

	fmt.Println("Primary API running on :8081")

	log.Fatal(http.ListenAndServe(":8081", nil))
}
