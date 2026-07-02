package auth

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/resend/resend-go/v3"
)

func RegisterRoutes(mux *http.ServeMux, service *Service, log *slog.Logger, mail *resend.Client) {
	h := NewHandler(service, log, mail)
	// mux.HandleFunc("GET /me", h.Me)
	fmt.Println(h)
	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", mux))
}
