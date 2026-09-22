package out

import (
	"context"

	"github.com/CunhazadanoDale/trads-market-test/internal/core/domain"
)

type CityRepository interface {
	Upsert(ctx context.Context, city *domain.City, stateIBGECode int64) error
	FindByState(
		ctx context.Context,
		stateIBGECode int64,
		page int,
		pageSize int,
	) ([]domain.CityWithIndicators, int, error)
	StateExists(ctx context.Context, stateIBGECode int64) (bool, error)
	FindDetailByIBGECode(ctx context.Context, ibgeCode int64) (*domain.CityDetail, error)
}
