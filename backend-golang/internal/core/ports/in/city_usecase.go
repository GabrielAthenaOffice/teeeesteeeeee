package in

import (
	"context"

	"github.com/CunhazadanoDale/trads-market-test/internal/core/domain"
)

type CityUseCase interface {
	Import(ctx context.Context) error
	FindByState(
		ctx context.Context,
		stateIBGECode int64,
		filter domain.PaginacaoFilter,
	) (domain.PaginacaoResponse[domain.City], error)
	FindByIBGECode(ctx context.Context, ibgeCode int64) (domain.CityDetail, error)
}
