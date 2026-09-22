package handler

import (
	"net/http"
	"strconv"

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

func (h *MetricsHandler) FindAgeDistribution(
	w http.ResponseWriter,
	r *http.Request,
) {
	distribution, err := h.useCase.FindAgeDistribution(r.Context())
	if err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			CodeInternalError,
			"falha ao buscar distribuição por faixa etária",
		)
		return
	}

	writeJSON(w, http.StatusOK, dtos.NewAgeDistributionResponse(distribution))
}

func (h *MetricsHandler) FindStates(
	w http.ResponseWriter,
	r *http.Request,
) {
	metrics, err := h.useCase.FindStates(r.Context())
	if err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			CodeInternalError,
			"falha ao buscar métricas por estado",
		)
		return
	}

	response := make([]dtos.StateMetricsResponse, 0, len(metrics))
	for _, item := range metrics {
		response = append(response, dtos.NewStateMetricsResponse(item))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *MetricsHandler) FindTopCities(
	w http.ResponseWriter,
	r *http.Request,
) {
	limit := 10

	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			writeError(
				w,
				http.StatusBadRequest,
				CodeInvalidRequest,
				"limit deve ser um número inteiro",
			)
			return
		}

		if value < 1 || value > 100 {
			writeError(
				w,
				http.StatusBadRequest,
				CodeInvalidRequest,
				"limit deve estar entre 1 e 100",
			)
			return
		}

		limit = value
	}

	top, err := h.useCase.FindTopCities(r.Context(), limit)
	if err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			CodeInternalError,
			"falha ao buscar ranking de municípios",
		)
		return
	}

	writeJSON(w, http.StatusOK, dtos.NewTopCitiesResponse(top))
}
