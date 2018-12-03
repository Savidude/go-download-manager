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
	// validate command args
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [url]...\n", os.Args[0])
		os.Exit(1)
	}

	urls := os.Args[1:]

	// start file downloads, 3 at a time
	fmt.Printf("Downloading %d files...\n", len(urls))
	start := time.Now()
	respch, err := go_download_manager.GetBatch(4, "./downloads", urls...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// monitor downloads
	responses := make([]*go_download_manager.Response, 0, len(urls))
	t := time.NewTicker(200 * time.Millisecond)
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
	// clear lines for incomplete downloads
	if inProgress > 0 {
		fmt.Printf("\033[%dA\033[K", inProgress)
	}

	// print newly completed downloads
	for i, resp := range responses {
		if resp != nil && resp.IsComplete() {
			if resp.Err() != nil {
				failed++
				fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n",
					resp.Request.URL(),
					resp.Err())
			} else {
				succeeded++
				fmt.Printf("Finished %s %d / %d bytes (%d%%)\n",
					resp.Filename,
					resp.BytesComplete(),
					resp.Size,
					int(100*resp.Progress()))
			}
			responses[i] = nil
		}
	}

	// print progress for incomplete downloads
	inProgress = 0
	for _, resp := range responses {
		if resp != nil {
			fmt.Printf("Downloading %s %d / %d bytes (%d%%) - %.02fKBp/s ETA: %ds \033[K\n",
				resp.Filename,
				resp.BytesComplete(),
				resp.Size,
				int(100*resp.Progress()),
				resp.BytesPerSecond()/1024,
				int64(resp.ETA().Sub(time.Now()).Seconds()))
			inProgress++
		}
	}
}
