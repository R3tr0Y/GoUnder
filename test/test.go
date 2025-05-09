package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

func main() {
	resp, err := http.DefaultClient.Get("https://r3tr0y.github.io")
	if err != nil {
		log.Fatal(err)
	}
	data, _ := io.ReadAll(resp.Body) // Ignoring error for example

	wappalyzerClient, err := wappalyzer.New()
	fingerprints := wappalyzerClient.Fingerprint(resp.Header, data)
	fmt.Printf("%v\n", fingerprints)

	// Output: map[Acquia Cloud Platform:{} Amazon EC2:{} Apache:{} Cloudflare:{} Drupal:{} PHP:{} Percona:{} React:{} Varnish:{}]
}
