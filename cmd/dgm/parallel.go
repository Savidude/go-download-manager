package main

import (
	"fmt"
	"github.com/Savidude/go-download-manager"
	"os"
	"time"
)

var (
	inProgress = 0
	failed     = 0
	succeeded  = 0
)

func main() {
	// get URL to download from command args
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s url\n", os.Args[0])
		os.Exit(1)
	}

	url := os.Args[1]
	var chunkSize = 1024 * 1024 * 10 //10MB

	// download file
	fmt.Printf("Downloading %s...\n", url)

	start := time.Now()
	respch, chunks, err := go_download_manager.GetParallel(".", url, int64(chunkSize), 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", url, err)
		os.Exit(1)
	}

	responses := make([]*go_download_manager.Response, 0, chunks)
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case resp := <-respch:
			if resp != nil {
				// a new response has been received and has started downloading
				responses = append(responses, resp)
			} else {
				// channel is closed - all downloads are complete
				updateUI(responses)
				break Loop
			}

		case <-t.C:
			// update UI every 200ms
			updateUI(responses)
		}
	}

	fmt.Printf(
		"Finished %d successful, %d failed, %d incomplete.\n",
		succeeded,
		failed,
		inProgress)
	elapsed := time.Since(start)
	fmt.Printf("Download took %s", elapsed)

	// return the number of failed downloads as exit code
	os.Exit(failed)
}

// updateUI prints the progress of all downloads to the terminal
func updateUI(responses []*go_download_manager.Response) {
	// print progress for incomplete downloads
	inProgress = 0
	var downloadedBytes int64 = 0
	var size int64 = 0
	for _, resp := range responses {
		downloadedBytes += resp.BytesComplete()
		size += resp.Size
	}

	if size != 0 {
		fmt.Printf("Downloading %d/%d bytes\n", downloadedBytes, size)
	}
}
