package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/CunhazadanoDale/trads-market-test/internal/adapter/http/dtos"
	"github.com/CunhazadanoDale/trads-market-test/internal/core/domain"
	"github.com/CunhazadanoDale/trads-market-test/internal/core/ports/in"
)

type CityHandler struct {
	useCase in.CityUseCase
}

func NewCityHandler(
	useCase in.CityUseCase,
) *CityHandler {
	return &CityHandler{
		useCase: useCase,
	}
}

func (h *CityHandler) FindByState(
	w http.ResponseWriter,
	r *http.Request,
) {
	stateIBGECode, err := strconv.ParseInt(
		r.PathValue("ibgeCode"),
		10,
		64,
	)
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			CodeInvalidRequest,
			"código IBGE do estado inválido",
		)
		return
	}

	filter := domain.PaginacaoFilter{
		Page: parseQueryInt(r, "page", 1),
		Size: parseQueryInt(r, "pageSize", 20),
	}

	result, err := h.useCase.FindByState(
		r.Context(),
		stateIBGECode,
		filter,
	)
	if err != nil {
		if errors.Is(err, domain.ErrStateNotFound) {
			writeError(
				w,
				http.StatusNotFound,
				CodeStateNotFound,
				fmt.Sprintf("estado com código IBGE %d não encontrado", stateIBGECode),
			)
			return
		}

		writeError(
			w,
			http.StatusInternalServerError,
			CodeInternalError,
			"falha ao buscar cidades",
		)
		return
	}

	response := domain.PaginacaoResponse[dtos.CityResponse]{
		Dados: make([]dtos.CityResponse, 0, len(result.Dados)),
		Page:  result.Page,
		Size:  result.Size,
		Total: result.Total,
	}

	for _, city := range result.Dados {
		response.Dados = append(response.Dados, dtos.NewCityResponse(city))
	}

	writeJSON(w, http.StatusOK, response)
}

func parseQueryInt(
	r *http.Request,
	name string,
	defaultValue int,
) int {
	value := r.URL.Query().Get(name)

	if value == "" {
		return defaultValue
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return result
}
