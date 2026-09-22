package out

import (
	"context"

	"github.com/CunhazadanoDale/trads-market-test/internal/core/domain"
)

type MetricsRepository interface {
	FindNational(ctx context.Context) (domain.NationalMetrics, error)
}
