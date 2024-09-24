package scan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"learning-go/tus-proto/structs"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

func Test(u string, wg *sync.WaitGroup) {
	defer wg.Done()
	var c structs.TargetUrl
	fmt.Println("[-] Scanning " + u)
	parsedUrl, err := url.Parse(u)
	if err != nil {
		fmt.Println("[!] Error Parsing URL " + u + "!")
		return
	}
	c.UrlProto = parsedUrl.Scheme
	host := parsedUrl.Host
	var domain, port string
	if strings.Contains(host, ":") {
		domain, port, err = net.SplitHostPort(host)
		if err != nil {
			fmt.Println("[!] Error Splitting Host and Port:", err)
			return
		}
	} else {
		domain = host
		if c.UrlProto == "http" {
			port = "80"
		} else if c.UrlProto == "https" {
			port = "443"
		}
	}
	c.Domain = domain
	c.Port = port
	resp, err := http.Get(u)
	if err != nil {
		fmt.Println("[!] " + u + " did not respond to a GET request!")
		return
	}
	defer resp.Body.Close()
	msg := fmt.Sprintf("[+] "+u+" responded with status code %d", resp.StatusCode)
	fmt.Println(msg)
	c.Code = int(resp.StatusCode)
	c.Protocol = resp.Proto
	c.Cookies = resp.Cookies()
	c.Headers = resp.Header
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[!] Could not read response body!")
	}
	c.Body = string(b)
	c.Url = u
	jsonData, err := json.Marshal(c)
	if err != nil {
		fmt.Println("Error Marshalling JSON: ", err)
		return
	}
	url := "http://localhost:8080/api/targeturl/new"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error Creating Request:", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	jsonResp, err := client.Do(req)
	if err != nil {
		fmt.Println("[!] Error Storing "+u+" in Database:", err)
	}
	defer jsonResp.Body.Close()
	fmt.Println("[+] Successfully Stored "+u+" in Database:", jsonResp.Status)
}
