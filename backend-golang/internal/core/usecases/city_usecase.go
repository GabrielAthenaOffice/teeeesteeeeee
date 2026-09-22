package usecases

import (
	"context"
	"fmt"

	"github.com/CunhazadanoDale/trads-market-test/internal/adapter/ibge"
	"github.com/CunhazadanoDale/trads-market-test/internal/core/domain"
	"github.com/CunhazadanoDale/trads-market-test/internal/core/ports/in"
	"github.com/CunhazadanoDale/trads-market-test/internal/core/ports/out"
)

var _ in.CityUseCase = (*CityUsecaseImpl)(nil)

type CityUsecaseImpl struct {
	repo       out.CityRepository
	ibgeClient *ibge.Client
}

func NewCityUsecaseImpl(repo out.CityRepository, ibgeClient *ibge.Client) *CityUsecaseImpl {
	return &CityUsecaseImpl{
		repo: repo,
		ibgeClient: ibgeClient,
	}
}

// Import implements [in.CityUseCase].
func (c *CityUsecaseImpl) Import(ctx context.Context) error {
	states, err := c.ibgeClient.GetStates(ctx)
	if err != nil {
		return fmt.Errorf("get states from IBGE : %w", err)
	}

	for _, state := range states {
		fmt.Printf("importando municipios de %s ... \n", state.Sigla)

		cities, err := c.ibgeClient.GetCitiesByState(ctx, state.Sigla)
		if err != nil {
			return fmt.Errorf("capturar cidades do estado %s : %w", state.Sigla, err)
		}

		for _, item := range cities {
			city := domain.City {
				IBGECode: item.ID,
				Name: item.Nome,
			}

			if err := c.repo.Upsert(ctx, &city, state.ID); err != nil {
				return fmt.Errorf("upsert city %d (%s) : %w",
				city.IBGECode, city.Name, err)
			}
		}

		fmt.Printf("%s: %d municípios importados\n",
			state.Sigla, len(cities),)
	}

	return nil
}

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// FindByState implements [in.CityUseCase].
func (c *CityUsecaseImpl) FindByState(
	ctx context.Context,
	stateIBGECode int64,
	filter domain.PaginacaoFilter,
) (domain.PaginacaoResponse[domain.City], error) {
	filter = normalizePaginacao(filter)

	cities, total, err := c.repo.FindByState(
		ctx,
		stateIBGECode,
		filter.Page,
		filter.Size,
	)
	if err != nil {
		return domain.PaginacaoResponse[domain.City]{},
			fmt.Errorf("buscar cidades do estado %d: %w", stateIBGECode, err)
	}

	if total == 0 {
		exists, err := c.repo.StateExists(ctx, stateIBGECode)
		if err != nil {
			return domain.PaginacaoResponse[domain.City]{},
				fmt.Errorf("verificar estado %d: %w", stateIBGECode, err)
		}

		if !exists {
			return domain.PaginacaoResponse[domain.City]{}, domain.ErrStateNotFound
		}
	}

	return domain.PaginacaoResponse[domain.City]{
		Dados: cities,
		Page:  filter.Page,
		Size:  filter.Size,
		Total: total,
	}, nil
}

// normalizePaginacao aplica os limites de paginacao:
// page < 1 → 1, size < 1 → 20, size > 100 → 100.
func normalizePaginacao(filter domain.PaginacaoFilter) domain.PaginacaoFilter {
	if filter.Page < defaultPage {
		filter.Page = defaultPage
	}

	if filter.Size < 1 {
		filter.Size = defaultPageSize
	}

	if filter.Size > maxPageSize {
		filter.Size = maxPageSize
	}

	return filter
}
