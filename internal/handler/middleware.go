package handler

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
)

type compressWriter struct {
	http.ResponseWriter
	Writer        io.Writer
	setEncoding   string
	setStatusCode int
}

func (w *compressWriter) Write(b []byte) (int, error) {
	contentType := w.Header().Get("Content-Type")
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
		w.Header().Set("Content-Encoding", w.setEncoding)
		w.ResponseWriter.WriteHeader(w.setStatusCode)
		return w.Writer.Write(b)
	}
	w.ResponseWriter.WriteHeader(w.setStatusCode)
	return w.ResponseWriter.Write(b)
}

func (w *compressWriter) WriteHeader(statusCode int) {
	w.setStatusCode = statusCode
}

func ResponseCompressHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoding := ""
		var writer io.Writer
		var err error = nil
		if strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
			wrb := brotli.NewWriterLevel(w, brotli.BestCompression)
			writer = wrb
			encoding = "br"
			defer wrb.Close()
		} else if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			wr, err1 := gzip.NewWriterLevel(w, gzip.BestCompression)
			err = err1
			writer = wr
			encoding = "gzip"
			defer wr.Close()
		} else if strings.Contains(r.Header.Get("Accept-Encoding"), "deflate") {
			wr, err1 := zlib.NewWriterLevel(w, flate.BestCompression)
			err = err1
			writer = wr
			encoding = "deflate"
			defer wr.Close()
		} else {
			next.ServeHTTP(w, r)
			return
		}

		if err != nil {
			io.WriteString(w, err.Error())
			return
		}

		next.ServeHTTP(&compressWriter{ResponseWriter: w, Writer: writer, setEncoding: encoding, setStatusCode: 0}, r)
	})
}

func RequestDecompressHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var reader io.ReadCloser
		var err error

		contentEncoding := r.Header.Get("Content-Encoding")

		if contentEncoding == `` {
			next.ServeHTTP(w, r)
			return
		}

		if contentEncoding == `gzip` {
			reader, err = gzip.NewReader(r.Body)
		} else if contentEncoding == `br` {
			reader = io.NopCloser(brotli.NewReader(r.Body))
		} else if contentEncoding == `deflate` {
			reader = flate.NewReader(r.Body)
		} else if contentEncoding != `` {
			http.Error(w, "decompressor not found", http.StatusInternalServerError)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		r.Body = reader
		defer reader.Close()
		r.Header.Del("Content-Encoding")
		next.ServeHTTP(w, r)
	})
}
