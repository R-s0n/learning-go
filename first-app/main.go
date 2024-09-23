package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
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

func builddatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatalf("Error Enabling Foreign Keys: %v", err)
	}
	createTargetUrlTable := `
	CREATE TABLE IF NOT EXISTS TargetUrl (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code INTEGER NOT NULL,
		protocol TEXT NOT NULL,
		body TEXT
	);`
	_, err = db.Exec(createTargetUrlTable)
	if err != nil {
		log.Fatalf("Error Creating TargetUrl Table: %v", err)
	}
	createCookiesTable := `
	CREATE TABLE IF NOT EXISTS Cookies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target_url_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		value TEXT NOT NULL,
		domain TEXT,
		path TEXT,
		expires DATETIME,
		secure BOOLEAN,
		http_only BOOLEAN,
		same_site TEXT,
		FOREIGN KEY (target_url_id) REFERENCES TargetUrl(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(createCookiesTable)
	if err != nil {
		log.Fatalf("Error Creating Cookies Table: %v", err)
	}
	createHeadersTable := `
	CREATE TABLE IF NOT EXISTS Headers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target_url_id INTEGER NOT NULL,
		header_name TEXT NOT NULL,
		header_value TEXT NOT NULL,
		FOREIGN KEY (target_url_id) REFERENCES TargetUrl(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(createHeadersTable)
	if err != nil {
		log.Fatalf("Error Creating Cookies Table: %v", err)
	}
	return db
}

func insertTargetUrl(db *sql.DB, url TargetUrl) error {
	targetUrlQuery := "INSERT INTO TargetUrl (code, protocol, body) VALUES (?, ?, ?)"
	result, err := db.Exec(targetUrlQuery, url.Code, url.Protocol, url.Body)
	if err != nil {
		return err
	}
	targetUrlID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	for _, cookie := range url.Cookies {
		cookieQuery := "INSERT INTO Cookies (target_url_id, name, value, domain, path, expires, secure, http_only, same_site) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
		_, err = db.Exec(cookieQuery, targetUrlID, cookie.Name, cookie.Value, cookie.Domain, cookie.Path, cookie.Expires, cookie.Secure, cookie.HttpOnly, cookie.SameSite)
		if err != nil {
			return err
		}
	}
	for key, values := range url.Headers {
		for _, value := range values {
			headerQuery := "INSERT INTO Headers (target_url_id, header_name, header_value) VALUES (?, ?, ?)"
			_, err = db.Exec(headerQuery, targetUrlID, key, value)
			if err != nil {
				return err
			}
		}
	}
	return nil
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

func createTargetUrl(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
			return
		}
		var url TargetUrl
		err := json.NewDecoder(r.Body).Decode(&url)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		err = insertTargetUrl(db, url)
		if err != nil {
			http.Error(w, "Error Inserting URL Into Database", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Target URL Inserted Successfully"))
	}
}

func handleRequests(db *sql.DB) {
	http.Handle("/helloworld", http.HandlerFunc(helloworld))
	http.Handle("/helloworldjson", http.HandlerFunc(helloworldjson))
	http.Handle("/testjson", http.HandlerFunc(testjson))
	http.HandleFunc("/api/targeturl/new", createTargetUrl(db))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	db := builddatabase()
	defer db.Close()
	handleRequests(db)
}
