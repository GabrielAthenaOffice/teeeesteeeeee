package usecases

import (
	"context"

	"github.com/CunhazadanoDale/trads-market-test/internal/core/domain"
	"github.com/CunhazadanoDale/trads-market-test/internal/core/ports/in"
	"github.com/CunhazadanoDale/trads-market-test/internal/core/ports/out"
)

var _ in.MetricsUseCase = (*MetricsUseCaseImpl)(nil)

type MetricsUseCaseImpl struct {
	repo out.MetricsRepository
}

func NewMetricsUseCaseImpl(repo out.MetricsRepository) *MetricsUseCaseImpl {
	return &MetricsUseCaseImpl{
		repo: repo,
	}
}

func (m *MetricsUseCaseImpl) FindNational(
	ctx context.Context,
) (domain.NationalMetrics, error) {
	return m.repo.FindNational(ctx)
}
