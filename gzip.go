package gzip

import (
	"compress/gzip"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

const (
	BestCompression    = gzip.BestCompression
	BestSpeed          = gzip.BestSpeed
	DefaultCompression = gzip.DefaultCompression
	NoCompression      = gzip.NoCompression
)

func Gzip(level int, options ...Option) gin.HandlerFunc {
	return newGzipHandler(level, options...).Handle
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) reset(r gin.ResponseWriter) {
	g.ResponseWriter = r
	g.writer.Reset(g.ResponseWriter)
}

func (g *gzipWriter) resetGZipWriter() {
	g.writer.Reset(g.ResponseWriter)
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	g.Header().Del("Content-Length")
	return g.writer.Write([]byte(s))
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	g.Header().Del("Content-Length")
	g.ResponseWriter.Header().Set("Content-Encoding", "gzip") // must before write
	count, err := g.writer.Write(data)
	_ = g.writer.Close() // must before reset
	g.writer.Reset(ioutil.Discard)
	putGzipWriter(g)
	return count, err
}

// Fix: https://github.com/mholt/caddy/issues/38
func (g *gzipWriter) WriteHeader(code int) {
	g.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(code)
}
