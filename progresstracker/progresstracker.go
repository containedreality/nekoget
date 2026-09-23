package progresstracker

import (
	"fmt"
	"io"
	"os"
)

type ProgressTracker struct {
	Underlying io.Writer
	Total      int64
	Current    int64
}

func (w *ProgressTracker) Write(p []byte) (int, error) {
	n, err := w.Underlying.Write(p)

	w.Current += int64(n)

	if w.Total < 1048576/(1024*1024) {
		fmt.Fprintf(os.Stderr, "\r%d / %db", w.Current, w.Total)
	} else {
		fmt.Fprintf(os.Stderr, "\r%fMiB / %fMiB", float64(w.Current)/(1024*1024), float64(w.Total)/(1024*1024))
	}

	return n, err
}
