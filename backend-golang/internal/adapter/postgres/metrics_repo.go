package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/CunhazadanoDale/trads-market-test/internal/core/domain"
	"github.com/CunhazadanoDale/trads-market-test/internal/core/ports/out"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ out.MetricsRepository = (*MetricsRepo)(nil)

type MetricsRepo struct {
	db *pgxpool.Pool
}

func NewMetricsRepository(db *pgxpool.Pool) *MetricsRepo {
	return &MetricsRepo{
		db: db,
	}
}

func (m *MetricsRepo) FindNational(
	ctx context.Context,
) (domain.NationalMetrics, error) {
	var metrics domain.NationalMetrics

	const municipalitiesQuery = `SELECT COUNT(*) FROM cities`

	if err := m.db.QueryRow(
		ctx,
		municipalitiesQuery,
	).Scan(&metrics.Municipalities); err != nil {
		return domain.NationalMetrics{}, fmt.Errorf(
			"count municipalities: %w",
			err,
		)
	}

	population, err := m.findPopulationTotal(ctx)
	if err != nil {
		return domain.NationalMetrics{}, err
	}
	metrics.Population = population

	income, err := m.findIncomeWeighted(ctx)
	if err != nil {
		return domain.NationalMetrics{}, err
	}
	metrics.Income = income

	gdp, err := m.findGDPTotal(ctx)
	if err != nil {
		return domain.NationalMetrics{}, err
	}
	metrics.GDP = gdp

	return metrics, nil
}

func (m *MetricsRepo) findPopulationTotal(
	ctx context.Context,
) (*domain.Indicator[int64], error) {
	const query = `
		SELECT year, SUM(value) AS total
		FROM population_indicators
		WHERE year = (SELECT MAX(year) FROM population_indicators)
		GROUP BY year
	`

	var year int
	var total int64

	if err := m.db.QueryRow(ctx, query).Scan(&year, &total); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query national population: %w", err)
	}

	return &domain.Indicator[int64]{Year: year, Value: total}, nil
}

func (m *MetricsRepo) findIncomeWeighted(
	ctx context.Context,
) (*domain.Indicator[float64], error) {
	const query = `
		SELECT
			i.year,
			ROUND((SUM(i.average_income * p.value) / SUM(p.value))::numeric, 2) AS weighted
		FROM income_indicators i
		INNER JOIN population_indicators p ON p.city_id = i.city_id
		WHERE i.year = (SELECT MAX(year) FROM income_indicators)
			AND p.year = (SELECT MAX(year) FROM population_indicators)
		GROUP BY i.year
		HAVING SUM(p.value) > 0
	`

	var year int
	var weighted float64

	if err := m.db.QueryRow(ctx, query).Scan(&year, &weighted); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query national income: %w", err)
	}

	return &domain.Indicator[float64]{Year: year, Value: weighted}, nil
}

func (m *MetricsRepo) findGDPTotal(
	ctx context.Context,
) (*domain.Indicator[float64], error) {
	const query = `
		SELECT year, SUM(gdp) AS total
		FROM gdp_indicators
		WHERE year = (SELECT MAX(year) FROM gdp_indicators)
		GROUP BY year
	`

	var year int
	var total float64

	if err := m.db.QueryRow(ctx, query).Scan(&year, &total); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query national gdp: %w", err)
	}

	return &domain.Indicator[float64]{Year: year, Value: total}, nil
}
