package dashboard

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
)

// Bodies below this size are not worth compressing.
const gzipMinBytes = 1024

var gzipPool = sync.Pool{New: func() interface{} {
	w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
	return w
}}

func gzipBytes(data []byte, level int) []byte {
	var buf bytes.Buffer
	w, _ := gzip.NewWriterLevel(&buf, level)
	_, _ = w.Write(data)
	_ = w.Close()
	return buf.Bytes()
}

// acceptsGzip reads an Accept-Encoding header.
func acceptsGzip(acceptEncoding string) bool {
	return strings.Contains(acceptEncoding, "gzip")
}

// compresses reports whether a body of size bytes is sent gzipped.
func compresses(size int, acceptEncoding string) bool {
	return size >= gzipMinBytes && acceptsGzip(acceptEncoding)
}

func writeJSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	body, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "application/json")
	h.Set("Cache-Control", "no-store")
	h.Set("Vary", "Accept-Encoding")
	if !compresses(len(body), r.Header.Get("Accept-Encoding")) {
		_, _ = w.Write(body)
		return
	}
	h.Set("Content-Encoding", "gzip")
	gz := gzipPool.Get().(*gzip.Writer)
	gz.Reset(w)
	_, _ = gz.Write(body)
	_ = gz.Close()
	gzipPool.Put(gz)
}
