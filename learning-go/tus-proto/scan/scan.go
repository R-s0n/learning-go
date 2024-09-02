package scan

import (
	"fmt"
	"net/http"
	"sync"
)

func Test(u string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("[-] Scanning " + u)
	resp, err := http.Get("http://example.com/")
	if err != nil {
		fmt.Println("[!] " + u + " did not respond to a GET request!")
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	fmt.Println(resp.StatusCode)
}
