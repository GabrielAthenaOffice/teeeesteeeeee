package main

import (
	"context"
	"log"
	"time"

	"github.com/CunhazadanoDale/trads-market-test/internal/adapter/ibge"
	"github.com/CunhazadanoDale/trads-market-test/internal/adapter/postgres"
	"github.com/CunhazadanoDale/trads-market-test/internal/config"
	"github.com/CunhazadanoDale/trads-market-test/internal/core/usecases"
)

func main() {
	cfg := config.LoadConfig()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		12*time.Minute,
	)
	defer cancel()

	db, err := postgres.NewDBConnection(ctx, cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("falha ao conectar com o database: %v", err)
	}
	defer db.Close()

	ibgeClient := ibge.NewIbgeClient(cfg.BaseUrlIBGE, cfg.BaseUrlLocalidades, nil)
	stateRepository := postgres.NewStateRepo(db)
	stateService := usecases.NewStateUseCase(stateRepository, ibgeClient)

	if err := stateService.Import(ctx); err != nil {
		log.Fatalf("falha ao importar estados: %v", err)
	}

	log.Println("states imported successfully")

	cityRepository := postgres.NewCityRepository(db)
	cityService := usecases.NewCityUsecaseImpl(cityRepository, ibgeClient)

	if err := cityService.Import(ctx); err != nil {
		log.Fatal("falha ao importar cidades: %w", err)
	}

	log.Println("cities imported successfully")

	populationRepository := postgres.NewPopulationRepo(db)
	populationService := usecases.NewPopulationUsecaseImpl(populationRepository, ibgeClient)

	if err := populationService.Import2022(ctx); err != nil {
		log.Fatalf("falha ao importar população: %v", err)
	}

	log.Println("population imported succesffuly")

	incomeRepository := postgres.NewIncomeRepo(db)
	incomeService := usecases.NewIncomeUsecaseImpl(incomeRepository, ibgeClient)

	if err := incomeService.Import(ctx); err != nil {
		log.Fatalf("falha ao importar renda: %v", err)
	}

	log.Println("income imported successfully")

	gdpRepository := postgres.NewGDPRepo(db)
	gdpUseCase := usecases.NewGDPUsecaseImpl(
		gdpRepository,
		ibgeClient,
	)

	if err := gdpUseCase.Import(ctx); err != nil {
		log.Fatalf("falha ao importar PIB: %v", err)
	}

	log.Println("GDP imported successfully")

	ageRepository := postgres.NewAgeRepo(db)
	ageUsecase := usecases.NewAgeUsecaseImpl(ageRepository, ibgeClient)

	if err := ageUsecase.Import(ctx); err != nil {
		log.Fatalf("falha ao importar faixa etária: %v", err)
	}

	log.Println("age imported successfully")

	log.Println("IBGE import successfully")
}