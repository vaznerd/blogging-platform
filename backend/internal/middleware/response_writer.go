package middleware

import "net/http"

type ResponseWriter struct {
	http.ResponseWriter
	Status int
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.Status = code
	rw.ResponseWriter.WriteHeader(code)
}
