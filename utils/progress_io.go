package utils

import (
	"io"
	"sync"
	"time"
)

type IOProgress struct {
	n         float64
	size      float64
	started   time.Time
	estimated time.Time
	err       error
}

type Reader struct {
	reader   io.Reader
	lock     sync.RWMutex
	Progress IOProgress
}

type Writer struct {
	writer   io.Writer
	lock     sync.RWMutex
	Progress IOProgress
}

func ReaderWithProgress(r io.Reader, size int64) *Reader {
	return &Reader{
		reader:   r,
		Progress: IOProgress{started: time.Now(), size: float64(size)},
	}
}

func WriterWithProgress(w io.Writer, size int64) *Writer {
	return &Writer{
		writer:   w,
		Progress: IOProgress{started: time.Now(), size: float64(size)},
	}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.lock.Lock()
	r.Progress.n += float64(n)
	r.Progress.err = err
	r.lock.Unlock()
	return n, err
}

func (w *Writer) Write(p []byte) (n int, err error) {
	n, err = w.writer.Write(p)
	w.lock.Lock()
	w.Progress.n += float64(n)
	w.Progress.err = err
	w.lock.Unlock()
	return n, err
}

func (p IOProgress) Size() float64 {
	return p.size
}

func (p IOProgress) N() float64 {
	return p.n
}

func (p IOProgress) Complete() bool {
	if p.err == io.EOF {
		return true
	}
	if p.size == -1 {
		return false
	}
	return p.n >= p.size
}

// Percent calculates the percentage complete.
func (p IOProgress) Percent() float64 {
	if p.n == 0 {
		return 0
	}
	if p.n >= p.size {
		return 100
	}
	return 100.0 / (p.size / p.n)
}

func (p IOProgress) Remaining() time.Duration {
	if p.estimated.IsZero() {
		return time.Until(p.Estimated())
	}
	return time.Until(p.estimated)
}

func (p IOProgress) Estimated() time.Time {
	ratio := p.n / p.size
	past := float64(time.Since(p.started))
	if p.n > 0.0 {
		total := time.Duration(past / ratio)
		p.estimated = p.started.Add(total)
	}
	return p.estimated
}
