package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"os"
	"strings"
)

// getHref goes though the Token's attributes until an "href" is found
// "bare" return will return the variables (ok, href) as defined in
// the function definition
func getHref(t html.Token) (ok bool, href string) {

	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}

	}

	return
}


// crawl, crawls webpages for links
// Extract all http** links from a given webpage
func crawl(url string, ch chan string, chFinished chan bool) {

  resp, err := http.Get(url)

	defer func() {
		chFinished <- true // Notice when the function is finished.
	}()

	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + url + "\"")
		return
	}

	b := resp.Body
	defer b.Close() // close Body when the function returns

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			return 			// End of the document, we're done
		case tt == html.StartTagToken:
			t := z.Token()


			isAnchor := t.Data == "a" // Check if the token is an <a> tag
			if !isAnchor {
				continue
			}

			ok, url := getHref(t) // Extract the href value, if there is one
			if !ok {
				continue
			}


			hasProto := strings.Index(url, "http") == 0 // Make sure the url begines with at least http**
			if hasProto {
				ch <- url
			}
		}
	}
}

func main() {
	foundUrls := make(map[string]bool)
	seedUrls := os.Args[1:]

	chUrls := make(chan string)
	chFinished := make(chan bool)


	for _, url := range seedUrls {
		go crawl(url, chUrls, chFinished) 	// Kick off the crawl process (concurrently)
	}

	// Subscribe to both channels
	for c := 0; c < len(seedUrls); {
		select {
		case url := <-chUrls:
			foundUrls[url] = true
		case <-chFinished:
			c++
		}
	}

	// We're done! Print the results...

	fmt.Println("\n Got the following Urls (", len(foundUrls), "):\n")

	for url, _ := range foundUrls {
		fmt.Println(" - " + url)
	}

	close(chUrls)
}
