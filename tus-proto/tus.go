package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func main() {
	fmt.Println("[+] Starting TUS...")
	if len(os.Args) == 2 {
		file := os.Args[1]
		fmt.Println("[-] Accessing " + file + "...")
		bs, err := os.ReadFile(file)
		if err != nil {
			fmt.Println("[!] Unable to access " + file + "!")
			return
		}
		str := string(bs)
		urls := strings.Split(str, "\n")
		validUrlArray := []string{}
		for _, u := range urls {
			_, err := url.ParseRequestURI(u)
			if err != nil {
				fmt.Println("[!] " + u + " is NOT a valid URL!  Skipping...")
				continue
			}
			fmt.Println("[+] Valid URL: " + u)
			validUrlArray = append(validUrlArray, u)
		}
		urlNumber := len(validUrlArray)
		output := fmt.Sprintf("[+] %d Valid URLs Found!", urlNumber)
		fmt.Println(output)
		for _, u := range validUrlArray {
			fmt.Println("[-] Scanning " + u)
		}
	} else {
		fmt.Println("[!] Please provide only a filename!\n[!] Example: ./tus file.ext")
	}
}
