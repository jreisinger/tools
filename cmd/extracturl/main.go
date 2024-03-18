package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/html"
)

func main() {
	log.SetPrefix(os.Args[0] + ": ")
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatal("please supply URL")
	}

	url := os.Args[1]

	urls, err := extractURLs(url)
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range urls {
		fmt.Println(u)
	}
}

func extractURLs(url string) ([]string, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	tokenizer := html.NewTokenizer(response.Body)
	var urls []string

	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			break
		}

		token := tokenizer.Token()

		if tokenType == html.StartTagToken && token.Data == "a" {
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					urls = append(urls, attr.Val)
				}
			}
		}
	}

	return urls, nil
}
