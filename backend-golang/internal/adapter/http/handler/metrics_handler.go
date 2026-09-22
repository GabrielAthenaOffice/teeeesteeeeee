package handler

import (
	"net/http"

	"github.com/CunhazadanoDale/trads-market-test/internal/adapter/http/dtos"
	"github.com/CunhazadanoDale/trads-market-test/internal/core/ports/in"
)

type MetricsHandler struct {
	useCase in.MetricsUseCase
}

func NewMetricsHandler(
	useCase in.MetricsUseCase,
) *MetricsHandler {
	return &MetricsHandler{
		useCase: useCase,
	}
}

func (h *MetricsHandler) FindNational(
	w http.ResponseWriter,
	r *http.Request,
) {
	metrics, err := h.useCase.FindNational(r.Context())
	if err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			CodeInternalError,
			"falha ao buscar métricas nacionais",
		)
		return
	}

	writeJSON(w, http.StatusOK, dtos.NewNationalMetricsResponse(metrics))
}
