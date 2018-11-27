package go_download_manager

import (
	"context"
	"hash"
	"net/http"
	"os"
	"time"
)

// checksum returns a hash of the given file, using the given hash algorithm.
func checksum(ctx context.Context, filename string, h hash.Hash) (b []byte, err error) {
	var f *os.File
	f, err = os.Open(filename)
	if err != nil {
		return
	}
	defer func() {
		err = f.Close()
	}()

	t := newTransfer(ctx, nil, h, f, nil)
	if _, err = t.copy(); err != nil {
		return
	}

	b = h.Sum(nil)
	return
}

// setLastModified sets the last modified timestamp of a local file according to
// the Last-Modified header returned by a remote server.
func setLastModified(resp *http.Response, filename string) error {
	// https://tools.ietf.org/html/rfc7232#section-2.2
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Last-Modified
	header := resp.Header.Get("Last-Modified")
	if header == "" {
		return nil
	}
	lastmod, err := time.Parse(http.TimeFormat, header)
	if err != nil {
		return nil
	}
	return os.Chtimes(filename, lastmod, lastmod)
}
