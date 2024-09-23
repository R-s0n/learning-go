package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type HelloMessage struct {
	Message string `json:"message"`
}

type TargetUrl struct {
	Code     int            `json:"code"`
	Protocol string         `json:"protocol"`
	Cookies  []*http.Cookie `json:"cookies"`
	Headers  http.Header    `json:"headers"`
	Body     string         `json:"body"`
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

func testjson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported.", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body.", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	var targetUrl TargetUrl
	err = json.Unmarshal(body, &targetUrl)
	if err != nil {
		http.Error(w, "Error parsing JSON.", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Received: %+v\n", targetUrl)
	jsonBodyRequest := fmt.Sprintf("Received: %+v", targetUrl)
	fmt.Println(jsonBodyRequest)
}

func handleRequests() {
	http.Handle("/helloworld", http.HandlerFunc(helloworld))
	http.Handle("/helloworldjson", http.HandlerFunc(helloworldjson))
	http.Handle("/testjson", http.HandlerFunc(testjson))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	handleRequests()
}
