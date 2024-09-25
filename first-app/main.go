package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type TargetUrl struct {
	ID       int            `json:"id"`
	Url      string         `json:"url"`
	UrlProto string         `json:"urlproto"`
	Domain   string         `json:"domain"`
	Port     string         `json:"port"`
	Code     int            `json:"code"`
	Protocol string         `json:"protocol"`
	Cookies  []*http.Cookie `json:"cookies"`
	Headers  http.Header    `json:"headers"`
	Body     string         `json:"body"`
}

type QueryUrl struct {
	Url string `json:"url"`
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
		url TEXT NOT NULL,
		url_proto TEXT NOT NULL,
		domain TEXT NOT NULL,
		port TEXT NOT NULL,
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
	targetUrlQuery := "INSERT INTO TargetUrl (url, url_proto, domain, port, code, protocol, body) VALUES (?, ?, ?, ?, ?, ?, ?)"
	result, err := db.Exec(targetUrlQuery, url.Url, url.UrlProto, url.Domain, url.Port, url.Code, url.Protocol, url.Body)
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

func urlExists(db *sql.DB, url string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM TargetUrl WHERE Url = ? LIMIT 1)"
	err := db.QueryRow(query, url).Scan(&exists)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	return exists, nil
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
		exists, err := urlExists(db, url.Url)
		if err != nil {
			http.Error(w, "Error Checking For Unique URL", http.StatusInternalServerError)
			return
		}
		if exists {
			http.Error(w, "URL Already Exists", http.StatusConflict)
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

func queryCookies(db *sql.DB, targetUrlId int) ([]*http.Cookie, error) {
	query := "SELECT name, value, domain, path, expires, secure, http_only, same_site FROM Cookies where target_url_id = ?"
	rows, err := db.Query(query, targetUrlId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cookies []*http.Cookie
	for rows.Next() {
		var cookie http.Cookie
		var expires string
		err := rows.Scan(&cookie.Name, &cookie.Value, &cookie.Domain, &cookie.Path, &expires, &cookie.Secure, &cookie.HttpOnly, &cookie.SameSite)
		if err != nil {
			return nil, err
		}
		cookies = append(cookies, &cookie)
	}
	return cookies, nil
}

func queryHeaders(db *sql.DB, targetUrlId int) (http.Header, error) {
	query := "SELECT header_name, header_value FROM Headers WHERE target_url_id = ?"
	rows, err := db.Query(query, targetUrlId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	headers := http.Header{}
	for rows.Next() {
		var headerName, headerValue string
		err := rows.Scan(&headerName, &headerValue)
		if err != nil {
			return nil, err
		}
		headers.Add(headerName, headerValue)
	}
	return headers, nil
}

func readTargetUrl(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
			return
		}
		var url TargetUrl
		var qurl QueryUrl
		err := json.NewDecoder(r.Body).Decode(&qurl)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		query := "SELECT id, url, url_proto, domain, port, code, protocol, body FROM TargetUrl WHERE url = ?"
		row := db.QueryRow(query, qurl.Url)
		err = row.Scan(&url.ID, &url.Url, &url.UrlProto, &url.Domain, &url.Port, &url.Code, &url.Protocol, &url.Body)
		if err == sql.ErrNoRows {
			fmt.Println("No TargetUrl Found")
			return
		} else if err != nil {
			fmt.Println("Something went wrong:", err)
			return
		}
		url.Cookies, err = queryCookies(db, url.ID)
		if err != nil {
			http.Error(w, "Error Retreiving Cookies", http.StatusInternalServerError)
			return
		}
		url.Headers, err = queryHeaders(db, url.ID)
		if err != nil {
			http.Error(w, "Error Retreiving Headers", http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(url)
	}
}

func updateCookies(db *sql.DB, targetUrlId int, cookies []*http.Cookie) error {
	_, err := db.Exec("DELETE FROM Cookies WHERE target_url_id = ?", targetUrlId)
	if err != nil {
		return err
	}
	for _, cookie := range cookies {
		query := `
			INSERT INTO Cookies (target_url_id, name, value, domain, path, expires, secure, http_only, same_site)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err := db.Exec(query, targetUrlId, cookie.Name, cookie.Value, cookie.Domain, cookie.Path, cookie.Expires, cookie.Secure, cookie.HttpOnly, cookie.SameSite)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateHeaders(db *sql.DB, targetUrlId int, headers http.Header) error {
	_, err := db.Exec("DELETE FROM Headers WHERE target_url_id = ?", targetUrlId)
	if err != nil {
		return err
	}
	for key, values := range headers {
		for _, value := range values {
			_, err := db.Exec("INSERT INTO Headers (target_url_id, header_name, header_value) VALUES (?, ?, ?)", targetUrlId, key, value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func updateTargetUrl(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
			return
		}
		var url TargetUrl
		err := json.NewDecoder(r.Body).Decode(&url)
		if err != nil || len(url.Url) < 1 {
			http.Error(w, "Invalid JSON or URL", http.StatusBadRequest)
			return
		}
		query := `
			UPDATE TargetUrl
			SET url = ?, url_proto = ?, domain = ?, port = ?, 
			code = ?, protocol = ?, body = ?
			WHERE url = ?`
		_, err = db.Exec(query, url.Url, url.UrlProto, url.Domain, url.Port, url.Code, url.Protocol, url.Body, url.Url)
		if err != nil {
			http.Error(w, "Error Updating Target URL", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
		if url.Cookies != nil {
			err = updateCookies(db, url.ID, url.Cookies)
			if err != nil {
				http.Error(w, "Failed to Update Cookies", http.StatusInternalServerError)
				return
			}
		}
		if url.Headers != nil {
			err = updateHeaders(db, url.ID, url.Headers)
			if err != nil {
				http.Error(w, "Failed to Update Headers", http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Target URL Updated Successfully"))
	}
}

func deleteTargetUrl(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
		}
		var qurl QueryUrl
		err := json.NewDecoder(r.Body).Decode(&qurl)
		if err != nil || len(qurl.Url) < 1 {
			http.Error(w, "Invalid JSON or URL", http.StatusBadRequest)
			return
		}
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to Begin Transaction", http.StatusInternalServerError)
			return
		}
		_, err = tx.Exec("DELETE FROM Cookies WHERE target_url_id = (SELECT id FROM TargetUrl WHERE url = ?)", qurl.Url)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to Delete Cookies", http.StatusInternalServerError)
			return
		}
		_, err = tx.Exec("DELETE FROM Headers WHERE target_url_id = (SELECT id FROM TargetUrl WHERE url = ?)", qurl.Url)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to Delete Headers", http.StatusInternalServerError)
			return
		}
		result, err := tx.Exec("DELETE FROM TargetUrl WHERE url = ?", qurl.Url)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to Delete TargetUrl", http.StatusInternalServerError)
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			http.Error(w, "Error Checking Rows Modified", http.StatusInternalServerError)
			return
		}
		if rowsAffected == 0 {
			tx.Rollback()
			http.Error(w, "TargetUrl Not Found", http.StatusNotFound)
			return
		}
		err = tx.Commit()
		if err != nil {
			http.Error(w, "Failed to Commit Transaction", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Target URL %s Deleted Successfully", qurl.Url)))
	}
}

func handleRequests(db *sql.DB) {
	http.HandleFunc("/api/targeturl/new", createTargetUrl(db))
	http.HandleFunc("/api/targeturl", readTargetUrl(db))
	http.HandleFunc("/api/targeturl/update", updateTargetUrl(db))
	http.HandleFunc("/api/targeturl/delete", deleteTargetUrl(db))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	db := builddatabase()
	defer db.Close()
	handleRequests(db)
}
