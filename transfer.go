package go_download_manager

import (
	"context"
	"io"
	"sync/atomic"
)

type transfer struct {
	n      int64 // must be 64bit aligned on 386
	ctx    context.Context
	lim    RateLimiter
	writer io.Writer
	reader io.Reader
	b      []byte
}

func newTransfer(ctx context.Context, lim RateLimiter, dst io.Writer, src io.Reader, buf []byte) *transfer {
	return &transfer{
		ctx:    ctx,
		lim:    lim,
		writer: dst,
		reader: src,
		b:      buf,
	}
}

// copy behaves similarly to io.CopyBuffer except that it checks for cancellation
// of the given context.Context and reports progress in a thread-safe manner.
func (t *transfer) copy() (written int64, err error) {
	if t.b == nil {
		t.b = make([]byte, 32*1024)
	}
	for {
		select {
		case <-t.ctx.Done():
			err = t.ctx.Err()
			return
		default:
			// keep working
		}
		if t.lim != nil {
			err = t.lim.WaitN(t.ctx, len(t.b))
			if err != nil {
				return
			}
		}
		nr, er := t.reader.Read(t.b)
		if nr > 0 {
			nw, ew := t.writer.Write(t.b[0:nr])
			if nw > 0 {
				written += int64(nw)
				atomic.StoreInt64(&t.n, written)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

// N returns the number of bytes transferred.
func (t *transfer) N() (n int64) {
	if t == nil {
		return 0
	}
	n = atomic.LoadInt64(&t.n)
	return
}
