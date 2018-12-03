package main

import (
	"fmt"
	"github.com/Savidude/go-download-manager"
	"os"
	"time"
)

func main() {
	// get URL to download from command args
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s url\n", os.Args[0])
		os.Exit(1)
	}

	url := os.Args[1]

	// download file
	fmt.Printf("Downloading %s...\n", url)
	start := time.Now()
	resp, err := go_download_manager.Get(".", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", url, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully downloaded to %s\n", resp.Filename)
	elapsed := time.Since(start)
	fmt.Printf("Download took %s", elapsed)
}
