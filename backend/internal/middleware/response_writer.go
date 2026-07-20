package middleware

import "net/http"

type ResponseWriter struct {
	http.ResponseWriter

	Status int
	wrote  bool
}

func (rw *ResponseWriter) WriteHeader(code int) {
	if rw.wrote {
		return
	}
	rw.Status = code
	rw.wrote = true
	rw.ResponseWriter.WriteHeader(code)
}
