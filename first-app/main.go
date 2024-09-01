package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type HelloMessage struct {
	Message string `json:"message"`
}

func helloworld(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello World!")
	fmt.Println("Endpoint Hit: helloWorld")
}

func helloworldjson(w http.ResponseWriter, r *http.Request) {
	msg := HelloMessage{"Hello World!"}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode((msg))
	fmt.Println("Endpoint Hit: helloWorldJson")
}

func handleRequests() {
	http.Handle("/helloworld", http.HandlerFunc(helloworld))
	http.Handle("/helloworldjson", http.HandlerFunc(helloworldjson))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	handleRequests()
}
