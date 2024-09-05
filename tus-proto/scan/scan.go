package scan

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"
)

func Test(u string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("[-] Scanning " + u)
	resp, err := http.Get(u)
	if err != nil {
		fmt.Println("[!] " + u + " did not respond to a GET request!")
		// fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	msg := fmt.Sprintf("[+] "+u+" responded with status code %d", resp.StatusCode)
	fmt.Println(msg)
	fmt.Println(resp.Cookies())
	fmt.Println(resp.Proto)
	fmt.Println(reflect.TypeOf(resp.Header))
	/*
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("[!] Could not read response body!")
		}
		fmt.Println(string(b))
	*/
}
