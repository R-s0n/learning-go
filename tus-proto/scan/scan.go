package scan

import (
	"fmt"
	"io"
	"learning-go/tus-proto/structs"
	"net/http"
	"sync"
)

func Test(u string, wg *sync.WaitGroup) {
	defer wg.Done()
	var c structs.TargetUrl
	fmt.Println("[-] Scanning " + u)
	// host := strings.Split(u, "//")
	// result, err := net.LookupHost(host[1])
	// if err != nil {
	// 	fmt.Println(result)
	// }
	resp, err := http.Get(u)
	if err != nil {
		fmt.Println("[!] " + u + " did not respond to a GET request!")
		// fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	msg := fmt.Sprintf("[+] "+u+" responded with status code %d", resp.StatusCode)
	fmt.Println(msg)
	c.Code = int(resp.StatusCode)
	// fmt.Println(resp.Proto)
	c.Protocol = resp.Proto
	// fmt.Println(resp.Cookies())
	c.Cookies = resp.Cookies()
	// fmt.Println(resp.Header)
	c.Headers = resp.Header
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[!] Could not read response body!")
	}
	// fmt.Println(string(b))
	c.Body = string(b)
	fmt.Println(c)
}
