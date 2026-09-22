package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "OK"})
}

func HealthHandlerWithDBCheck(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			writeError(
				w,
				http.StatusInternalServerError,
				CodeInternalError,
				"falha ao conectar com o banco de dados",
			)
			return
		}

		writeJSON(w, http.StatusOK, HealthResponse{Status: "OK"})
	}
}
