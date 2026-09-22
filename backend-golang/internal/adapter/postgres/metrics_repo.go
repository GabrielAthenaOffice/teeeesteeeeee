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

func (m *MetricsRepo) FindStates(
	ctx context.Context,
) ([]domain.StateMetrics, error) {
	const query = `
		SELECT
			s.id,
			s.ibge_code,
			s.name,
			s.uf,
			s.region,
			COUNT(c.id) AS municipios
		FROM states s
		LEFT JOIN cities c ON c.state_id = s.id
		GROUP BY s.id, s.ibge_code, s.name, s.uf, s.region
		ORDER BY s.name
	`

	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query state metrics: %w", err)
	}
	defer rows.Close()

	states := make([]domain.StateMetrics, 0)

	for rows.Next() {
		var item domain.StateMetrics

		if err := rows.Scan(
			&item.State.ID,
			&item.State.IBGECode,
			&item.State.Name,
			&item.State.UF,
			&item.State.Region,
			&item.Municipalities,
		); err != nil {
			return nil, fmt.Errorf("scan state metrics: %w", err)
		}

		states = append(states, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate state metrics: %w", err)
	}

	populations, err := m.findPopulationsByStateIDs(ctx)
	if err != nil {
		return nil, err
	}

	incomes, err := m.findIncomesByStateIDs(ctx)
	if err != nil {
		return nil, err
	}

	gdps, err := m.findGDPsByStateIDs(ctx)
	if err != nil {
		return nil, err
	}

	for index := range states {
		states[index].Population = populations[states[index].State.ID]
		states[index].Income = incomes[states[index].State.ID]
		states[index].GDP = gdps[states[index].State.ID]
	}

	return states, nil
}

func (m *MetricsRepo) findPopulationsByStateIDs(
	ctx context.Context,
) (map[int64]*domain.Indicator[int64], error) {
	const query = `
		SELECT c.state_id, p.year, SUM(p.value) AS total
		FROM population_indicators p
		INNER JOIN cities c ON c.id = p.city_id
		WHERE p.year = (SELECT MAX(year) FROM population_indicators)
		GROUP BY c.state_id, p.year
	`

	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query state populations: %w", err)
	}
	defer rows.Close()

	indicators := make(map[int64]*domain.Indicator[int64])

	for rows.Next() {
		var stateID int64
		var year int
		var total int64

		if err := rows.Scan(&stateID, &year, &total); err != nil {
			return nil, fmt.Errorf("scan state population: %w", err)
		}

		indicators[stateID] = &domain.Indicator[int64]{Year: year, Value: total}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate state populations: %w", err)
	}

	return indicators, nil
}

func (m *MetricsRepo) findIncomesByStateIDs(
	ctx context.Context,
) (map[int64]*domain.Indicator[float64], error) {
	const query = `
		SELECT
			c.state_id,
			i.year,
			ROUND((SUM(i.average_income * p.value) / SUM(p.value))::numeric, 2) AS weighted
		FROM income_indicators i
		INNER JOIN population_indicators p ON p.city_id = i.city_id
		INNER JOIN cities c ON c.id = i.city_id
		WHERE i.year = (SELECT MAX(year) FROM income_indicators)
			AND p.year = (SELECT MAX(year) FROM population_indicators)
		GROUP BY c.state_id, i.year
		HAVING SUM(p.value) > 0
	`

	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query state incomes: %w", err)
	}
	defer rows.Close()

	indicators := make(map[int64]*domain.Indicator[float64])

	for rows.Next() {
		var stateID int64
		var year int
		var weighted float64

		if err := rows.Scan(&stateID, &year, &weighted); err != nil {
			return nil, fmt.Errorf("scan state income: %w", err)
		}

		indicators[stateID] = &domain.Indicator[float64]{Year: year, Value: weighted}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate state incomes: %w", err)
	}

	return indicators, nil
}

func (m *MetricsRepo) findGDPsByStateIDs(
	ctx context.Context,
) (map[int64]*domain.Indicator[float64], error) {
	const query = `
		SELECT c.state_id, g.year, SUM(g.gdp) AS total
		FROM gdp_indicators g
		INNER JOIN cities c ON c.id = g.city_id
		WHERE g.year = (SELECT MAX(year) FROM gdp_indicators)
		GROUP BY c.state_id, g.year
	`

	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query state gdps: %w", err)
	}
	defer rows.Close()

	indicators := make(map[int64]*domain.Indicator[float64])

	for rows.Next() {
		var stateID int64
		var year int
		var total float64

		if err := rows.Scan(&stateID, &year, &total); err != nil {
			return nil, fmt.Errorf("scan state gdp: %w", err)
		}

		indicators[stateID] = &domain.Indicator[float64]{Year: year, Value: total}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate state gdps: %w", err)
	}

	return indicators, nil
}

func (m *MetricsRepo) FindTopCities(
	ctx context.Context,
	limit int,
) (domain.TopCities, error) {
	populationCities, err := m.findTopPopulationCities(ctx, limit)
	if err != nil {
		return domain.TopCities{}, err
	}

	incomeCities, err := m.findTopIncomeCities(ctx, limit)
	if err != nil {
		return domain.TopCities{}, err
	}

	gdpCities, err := m.findTopGDPCities(ctx, limit)
	if err != nil {
		return domain.TopCities{}, err
	}

	cityIDs := make([]int64, 0, 3*limit)
	for _, cities := range [][]domain.CityDetail{populationCities, incomeCities, gdpCities} {
		for _, city := range cities {
			cityIDs = append(cityIDs, city.City.ID)
		}
	}

	populations, err := m.findPopulationsByCityIDs(ctx, cityIDs)
	if err != nil {
		return domain.TopCities{}, err
	}

	incomes, err := m.findIncomesByCityIDs(ctx, cityIDs)
	if err != nil {
		return domain.TopCities{}, err
	}

	gdps, err := m.findGDPsByCityIDs(ctx, cityIDs)
	if err != nil {
		return domain.TopCities{}, err
	}

	return domain.TopCities{
		TopPopulation: attachIndicators(populationCities, populations, incomes, gdps),
		TopIncome:     attachIndicators(incomeCities, populations, incomes, gdps),
		TopGDP:        attachIndicators(gdpCities, populations, incomes, gdps),
	}, nil
}

func (m *MetricsRepo) findTopPopulationCities(
	ctx context.Context,
	limit int,
) ([]domain.CityDetail, error) {
	const query = `
		SELECT
			c.id,
			c.ibge_code,
			c.name,
			c.state_id,
			s.id,
			s.ibge_code,
			s.name,
			s.uf,
			s.region
		FROM population_indicators p
		INNER JOIN cities c ON c.id = p.city_id
		INNER JOIN states s ON s.id = c.state_id
		WHERE p.year = (SELECT MAX(year) FROM population_indicators)
		ORDER BY p.value DESC, c.name ASC, c.ibge_code ASC
		LIMIT $1
	`

	return m.queryTopCities(ctx, query, limit)
}

func (m *MetricsRepo) findTopIncomeCities(
	ctx context.Context,
	limit int,
) ([]domain.CityDetail, error) {
	const query = `
		SELECT
			c.id,
			c.ibge_code,
			c.name,
			c.state_id,
			s.id,
			s.ibge_code,
			s.name,
			s.uf,
			s.region
		FROM income_indicators i
		INNER JOIN cities c ON c.id = i.city_id
		INNER JOIN states s ON s.id = c.state_id
		WHERE i.year = (SELECT MAX(year) FROM income_indicators)
		ORDER BY i.average_income DESC, c.name ASC, c.ibge_code ASC
		LIMIT $1
	`

	return m.queryTopCities(ctx, query, limit)
}

func (m *MetricsRepo) findTopGDPCities(
	ctx context.Context,
	limit int,
) ([]domain.CityDetail, error) {
	const query = `
		SELECT
			c.id,
			c.ibge_code,
			c.name,
			c.state_id,
			s.id,
			s.ibge_code,
			s.name,
			s.uf,
			s.region
		FROM gdp_indicators g
		INNER JOIN cities c ON c.id = g.city_id
		INNER JOIN states s ON s.id = c.state_id
		WHERE g.year = (SELECT MAX(year) FROM gdp_indicators)
		ORDER BY g.gdp DESC, c.name ASC, c.ibge_code ASC
		LIMIT $1
	`

	return m.queryTopCities(ctx, query, limit)
}

func (m *MetricsRepo) queryTopCities(
	ctx context.Context,
	query string,
	limit int,
) ([]domain.CityDetail, error) {
	rows, err := m.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query top cities: %w", err)
	}
	defer rows.Close()

	cities := make([]domain.CityDetail, 0)

	for rows.Next() {
		var city domain.CityDetail

		if err := rows.Scan(
			&city.City.ID,
			&city.City.IBGECode,
			&city.City.Name,
			&city.City.StateID,
			&city.State.ID,
			&city.State.IBGECode,
			&city.State.Name,
			&city.State.UF,
			&city.State.Region,
		); err != nil {
			return nil, fmt.Errorf("scan top city: %w", err)
		}

		cities = append(cities, city)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate top cities: %w", err)
	}

	return cities, nil
}

func attachIndicators(
	cities []domain.CityDetail,
	populations map[int64]*domain.Indicator[int64],
	incomes map[int64]*domain.Indicator[float64],
	gdps map[int64]*domain.Indicator[float64],
) []domain.CityDetail {
	for index := range cities {
		cityID := cities[index].City.ID
		cities[index].Population = populations[cityID]
		cities[index].Income = incomes[cityID]
		cities[index].GDP = gdps[cityID]
	}

	return cities
}

func (m *MetricsRepo) findPopulationsByCityIDs(
	ctx context.Context,
	cityIDs []int64,
) (map[int64]*domain.Indicator[int64], error) {
	const query = `
		SELECT DISTINCT ON (city_id) city_id, year, value
		FROM population_indicators
		WHERE city_id = ANY($1)
		ORDER BY city_id, year DESC
	`

	rows, err := m.db.Query(ctx, query, cityIDs)
	if err != nil {
		return nil, fmt.Errorf("query populations by city ids: %w", err)
	}
	defer rows.Close()

	indicators := make(map[int64]*domain.Indicator[int64])

	for rows.Next() {
		var cityID int64
		var year int
		var value int64

		if err := rows.Scan(&cityID, &year, &value); err != nil {
			return nil, fmt.Errorf("scan city population: %w", err)
		}

		indicators[cityID] = &domain.Indicator[int64]{Year: year, Value: value}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate city populations: %w", err)
	}

	return indicators, nil
}

func (m *MetricsRepo) findIncomesByCityIDs(
	ctx context.Context,
	cityIDs []int64,
) (map[int64]*domain.Indicator[float64], error) {
	const query = `
		SELECT DISTINCT ON (city_id) city_id, year, average_income
		FROM income_indicators
		WHERE city_id = ANY($1)
		ORDER BY city_id, year DESC
	`

	rows, err := m.db.Query(ctx, query, cityIDs)
	if err != nil {
		return nil, fmt.Errorf("query incomes by city ids: %w", err)
	}
	defer rows.Close()

	indicators := make(map[int64]*domain.Indicator[float64])

	for rows.Next() {
		var cityID int64
		var year int
		var value float64

		if err := rows.Scan(&cityID, &year, &value); err != nil {
			return nil, fmt.Errorf("scan city income: %w", err)
		}

		indicators[cityID] = &domain.Indicator[float64]{Year: year, Value: value}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate city incomes: %w", err)
	}

	return indicators, nil
}

func (m *MetricsRepo) findGDPsByCityIDs(
	ctx context.Context,
	cityIDs []int64,
) (map[int64]*domain.Indicator[float64], error) {
	const query = `
		SELECT DISTINCT ON (city_id) city_id, year, gdp
		FROM gdp_indicators
		WHERE city_id = ANY($1)
		ORDER BY city_id, year DESC
	`

	rows, err := m.db.Query(ctx, query, cityIDs)
	if err != nil {
		return nil, fmt.Errorf("query gdps by city ids: %w", err)
	}
	defer rows.Close()

	indicators := make(map[int64]*domain.Indicator[float64])

	for rows.Next() {
		var cityID int64
		var year int
		var value float64

		if err := rows.Scan(&cityID, &year, &value); err != nil {
			return nil, fmt.Errorf("scan city gdp: %w", err)
		}

		indicators[cityID] = &domain.Indicator[float64]{Year: year, Value: value}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate city gdps: %w", err)
	}

	return indicators, nil
}
