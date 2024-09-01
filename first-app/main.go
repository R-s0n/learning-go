package main

import (
	"fmt"
	"log"
	"net/http"
)

func helloworld(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello World!")
	fmt.Println("Endpoint Hit: helloWorld")
}

func handleRequests() {
	http.Handle("/helloworld", http.HandlerFunc(helloworld))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	handleRequests()
}
