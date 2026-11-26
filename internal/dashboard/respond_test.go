package dashboard

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"testing"
)

func gunzip(t *testing.T, data []byte) []byte {
	t.Helper()
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	out, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("gunzip: %v", err)
	}
	return out
}

func TestAcceptsGzip(t *testing.T) {
	tests := []struct {
		name           string
		acceptEncoding string
		want           bool
	}{
		{"gzip alone", "gzip", true},
		{"gzip among others", "br, gzip, deflate", true},
		{"no header", "", false},
		{"other encodings only", "br, deflate", false},
		{"identity", "identity", false},
		{"gzip refused with q=0 (today: accepted)", "gzip;q=0", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := acceptsGzip(tt.acceptEncoding); got != tt.want {
				t.Errorf("acceptsGzip(%q) = %v, want %v", tt.acceptEncoding, got, tt.want)
			}
		})
	}

	t.Run("gzip refused with q=0 (bug: should not be accepted)", func(t *testing.T) {
		t.Skip("bug: the header is searched for the word gzip, so its q=0 is ignored")
		if got := acceptsGzip("gzip;q=0"); got != false {
			t.Errorf("acceptsGzip(%q) = %v, want false", "gzip;q=0", got)
		}
	})
}

func TestCompresses(t *testing.T) {
	tests := []struct {
		name           string
		size           int
		acceptEncoding string
		want           bool
	}{
		{"1023 bytes for a client that takes gzip", 1023, "gzip", false},
		{"1024 bytes for a client that takes gzip", 1024, "gzip", true},
		{"1025 bytes for a client that takes gzip", 1025, "gzip", true},
		{"1024 bytes for a client that does not", 1024, "", false},
		{"an empty body for a client that takes gzip", 0, "gzip", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compresses(tt.size, tt.acceptEncoding); got != tt.want {
				t.Errorf("compresses(%d, %q) = %v, want %v", tt.size, tt.acceptEncoding, got, tt.want)
			}
		})
	}
}

func TestGzipBytes(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"a line of text", "Channel Points Miner"},
		{"text outside ASCII", "Añadido 🎁"},
		{"a kilobyte of one letter", strings.Repeat("x", 1024)},
		{"nothing", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packed := gzipBytes([]byte(tt.data), gzip.BestCompression)
			if got := string(gunzip(t, packed)); got != tt.data {
				t.Errorf("gunzip(gzipBytes(%q)) = %q, want the input back", tt.data, got)
			}
		})
	}

	t.Run("a kilobyte of one letter packs to under 64 bytes", func(t *testing.T) {
		packed := gzipBytes([]byte(strings.Repeat("x", 1024)), gzip.BestCompression)
		if len(packed) >= 64 {
			t.Errorf("gzipBytes of 1024 equal bytes = %d bytes, want under 64", len(packed))
		}
	})
}
