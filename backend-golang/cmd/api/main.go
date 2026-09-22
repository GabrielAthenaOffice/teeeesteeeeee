package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	appRouter "github.com/CunhazadanoDale/trads-market-test/internal/adapter/http"
	"github.com/CunhazadanoDale/trads-market-test/internal/adapter/ibge"
	"github.com/CunhazadanoDale/trads-market-test/internal/adapter/postgres"
	"github.com/CunhazadanoDale/trads-market-test/internal/config"
	"github.com/CunhazadanoDale/trads-market-test/internal/core/usecases"
)

func main() {
	cfg := config.LoadConfig()

	context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := postgres.NewDBConnection(context, cfg.DatabaseUrl)
	if err != nil {
		log.Fatal(fmt.Errorf("erro ao conectar ao database: %w", err))
	}
	defer db.Close()

	ibgeClient := ibge.NewIbgeClient(cfg.BaseUrlIBGE, cfg.BaseUrlLocalidades, http.DefaultClient)

	stateRepository := postgres.NewStateRepo(db)
	stateUseCase := usecases.NewStateUseCase(stateRepository, ibgeClient)

	cityRepository := postgres.NewCityRepository(db)
	cityUseCase := usecases.NewCityUsecaseImpl(cityRepository, ibgeClient)

	metricsRepository := postgres.NewMetricsRepository(db)
	metricsUseCase := usecases.NewMetricsUseCaseImpl(metricsRepository)

	router := appRouter.NewRouter(db, stateUseCase, cityUseCase, metricsUseCase)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	log.Printf("API rodando em %s, %s", server.Addr, cfg.AppEnv)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
