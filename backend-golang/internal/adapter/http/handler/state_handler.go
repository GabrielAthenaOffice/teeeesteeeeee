package handler

import (
	"net/http"

	"github.com/CunhazadanoDale/trads-market-test/internal/adapter/http/dtos"
	"github.com/CunhazadanoDale/trads-market-test/internal/core/ports/in"
)

type StateHandler struct {
	useCase in.StateUseCase
}

func NewStateHandler(
	useCase in.StateUseCase,
) *StateHandler {
	return &StateHandler{
		useCase: useCase,
	}
}

func (h *StateHandler) FindAll(
	w http.ResponseWriter,
	r *http.Request,
) {
	states, err := h.useCase.FindAll(r.Context())
	if err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			CodeInternalError,
			"falha ao buscar estados",
		)
		return
	}

	response := make([]dtos.StateResponse, 0, len(states))
	for _, state := range states {
		response = append(response, dtos.NewStateResponse(state))
	}

	writeJSON(w, http.StatusOK, response)
}
