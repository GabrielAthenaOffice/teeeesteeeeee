package http

import (
	"net/http"

	"github.com/CunhazadanoDale/trads-market-test/internal/adapter/http/handler"
	"github.com/CunhazadanoDale/trads-market-test/internal/core/ports/in"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(
	db *pgxpool.Pool,
	stateUseCase in.StateUseCase,
	cityUseCase in.CityUseCase,
	metricsUseCase in.MetricsUseCase,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.HealthHandler)
	mux.HandleFunc("GET /health/db", handler.HealthHandlerWithDBCheck(db))

	stateHandler := handler.NewStateHandler(stateUseCase)
	mux.HandleFunc("GET /api/v1/states", stateHandler.FindAll)

	cityHandler := handler.NewCityHandler(cityUseCase)
	mux.HandleFunc("GET /api/v1/states/{ibgeCode}/cities", cityHandler.FindByState)

	mux.HandleFunc("GET /api/v1/cities/{ibgeCode}", cityHandler.FindByIBGECode)

	metricsHandler := handler.NewMetricsHandler(metricsUseCase)
	mux.HandleFunc("GET /api/v1/dashboard/national", metricsHandler.FindNational)

	mux.HandleFunc("GET /", handler.NotFoundHandler)

	return mux
}
