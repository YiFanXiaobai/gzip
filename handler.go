package gzip

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var handler *gzipHandler

type gzipHandler struct {
	*Options
	gzipWriterPool sync.Pool
}

func newGzipHandler(level int, options ...Option) *gzipHandler {
	handler = &gzipHandler{
		Options: DefaultOptions,
		gzipWriterPool: sync.Pool{
			New: func() interface{} {
				gz, err := gzip.NewWriterLevel(ioutil.Discard, level)
				if err != nil {
					panic(err)
				}
				return &gzipWriter{ResponseWriter: nil, writer: gz}
			},
		},
	}
	for _, setter := range options {
		setter(handler.Options)
	}
	return handler
}

func (g *gzipHandler) Handle(c *gin.Context) {
	if fn := g.DecompressFn; fn != nil && c.Request.Header.Get("Content-Encoding") == "gzip" {
		fn(c)
	}

	if !g.shouldCompress(c.Request) {
		return
	}

	gzW := g.gzipWriterPool.Get().(*gzipWriter)
	gzW.reset(c.Writer)
	c.Writer = gzW

	c.Next()
}

func (g *gzipHandler) shouldCompress(req *http.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") ||
		strings.Contains(req.Header.Get("Connection"), "Upgrade") ||
		strings.Contains(req.Header.Get("Accept"), "text/event-stream") {
		return false
	}

	extension := filepath.Ext(req.URL.Path)
	if g.ExcludedExtensions.Contains(extension) {
		return false
	}

	if g.ExcludedPaths.Contains(req.URL.Path) {
		return false
	}
	if g.ExcludedPathesRegexs.Contains(req.URL.Path) {
		return false
	}

	return g.IncludedPathsRegexs.Contains(req.URL.Path)
}

func putGzipWriter(gz *gzipWriter) {
	handler.gzipWriterPool.Put(gz)
}
