package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/spf13/cobra"
)

var fingerprintCmd = &cobra.Command{
	Use:   "fingerprint",
	Short: "Analyze fingerprints of websites.",
	Run: func(cmd *cobra.Command, args []string) {
		if targetURL == "" {
			fmt.Println("❗ use -u for target URL")
			_ = cmd.Usage()
			os.Exit(1)
		}
		fingerprintLookup(targetURL)
	},
}

func fingerprintLookup(url string) {
	fmt.Println("Automatically using wappalyzer...")
	wappalyzerAnalyze(url)

}
func wappalyzerAnalyze(url string) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	data, _ := io.ReadAll(resp.Body) // Ignoring error for example

	wappalyzerClient, err := wappalyzer.New()
	if err != nil {
		log.Fatal(err)
	} else {
		fingerprints := wappalyzerClient.Fingerprint(resp.Header, data)
		if len(fingerprints) > 0 {
			fmt.Println("\n✅ Website fingerprints found:")
			for fingerprint := range fingerprints {
				fmt.Printf("%v\n", fingerprint)
			}
		} else {
			fmt.Println("No website fingerprints found!")
		}

	}
}
func init() {
	fingerprintCmd.Flags().StringVarP(&targetURL, "url", "u", "", "targetURL（eg: https://example.com）")
	rootCmd.AddCommand(fingerprintCmd)
}
