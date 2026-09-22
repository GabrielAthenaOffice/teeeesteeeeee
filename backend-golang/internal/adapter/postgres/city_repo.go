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

var _ out.CityRepository = (*CityRepo)(nil)

type CityRepo struct {
	db *pgxpool.Pool
}

func NewCityRepository(db *pgxpool.Pool) *CityRepo {
	return &CityRepo{
		db: db,
	}
}

// Upsert implements [out.CityRepository].
func (c *CityRepo) Upsert(ctx context.Context, city *domain.City, stateIBGECode int64) error {
	query := `INSERT INTO cities (
			ibge_code,
			state_id,
			name
		)
		SELECT
			$1,
			id,
			$2
		FROM states
		WHERE ibge_code = $3
		ON CONFLICT (ibge_code)
		DO UPDATE SET
			state_id = EXCLUDED.state_id,
			name = EXCLUDED.name,
			updated_at = NOW()
	`

	_, err := c.db.Exec(
		ctx, query, city.IBGECode, city.Name, stateIBGECode,
	)

	return err
}

func (c *CityRepo) StateExists(ctx context.Context, stateIBGECode int64) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM states WHERE ibge_code = $1)`

	var exists bool

	if err := c.db.QueryRow(
		ctx,
		query,
		stateIBGECode,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf(
			"check state %d: %w",
			stateIBGECode,
			err,
		)
	}

	return exists, nil
}

// FindByState implements [out.CityRepository].
func (c *CityRepo) FindByState(
	ctx context.Context,
	stateIBGECode int64,
	page int,
	pageSize int,
) ([]domain.City, int, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM cities c
		INNER JOIN states s ON s.id = c.state_id
		WHERE s.ibge_code = $1
	`

	var total int

	if err := c.db.QueryRow(
		ctx,
		countQuery,
		stateIBGECode,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf(
			"count cities by state: %w",
			err,
		)
	}

	const query = `
		SELECT
			c.id,
			c.ibge_code,
			c.name,
			c.state_id
		FROM cities c
		INNER JOIN states s ON s.id = c.state_id
		WHERE s.ibge_code = $1
		ORDER BY c.name
		LIMIT $2
		OFFSET $3
	`

	offset := (page - 1) * pageSize

	rows, err := c.db.Query(
		ctx,
		query,
		stateIBGECode,
		pageSize,
		offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf(
			"query cities by state: %w",
			err,
		)
	}
	defer rows.Close()

	cities := make([]domain.City, 0)

	for rows.Next() {
		var city domain.City

		if err := rows.Scan(
			&city.ID,
			&city.IBGECode,
			&city.Name,
			&city.StateID,
		); err != nil {
			return nil, 0, fmt.Errorf(
				"scan city: %w",
				err,
			)
		}

		cities = append(cities, city)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf(
			"iterate cities: %w",
			err,
		)
	}

	return cities, total, nil
}

func (c *CityRepo) FindDetailByIBGECode(ctx context.Context, ibgeCode int64) (*domain.CityDetail, error) {
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
		FROM cities c
		INNER JOIN states s ON s.id = c.state_id
		WHERE c.ibge_code = $1
	`

	var detail domain.CityDetail

	if err := c.db.QueryRow(ctx, query, ibgeCode).Scan(
		&detail.City.ID,
		&detail.City.IBGECode,
		&detail.City.Name,
		&detail.City.StateID,
		&detail.State.ID,
		&detail.State.IBGECode,
		&detail.State.Name,
		&detail.State.UF,
		&detail.State.Region,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCityNotFound
		}
		return nil, fmt.Errorf("query city detail %d: %w", ibgeCode, err)
	}

	population, err := c.findPopulation(ctx, detail.City.ID)
	if err != nil {
		return nil, err
	}
	detail.Population = population

	income, err := c.findIncome(ctx, detail.City.ID)
	if err != nil {
		return nil, err
	}
	detail.Income = income

	gdp, err := c.findGDP(ctx, detail.City.ID)
	if err != nil {
		return nil, err
	}
	detail.GDP = gdp

	return &detail, nil
}

func (c *CityRepo) findPopulation(ctx context.Context, cityID int64) (*domain.Indicator[int64], error) {
	const query = `
		SELECT year, value
		FROM population_indicators
		WHERE city_id = $1
		ORDER BY year DESC
		LIMIT 1
	`

	var year int
	var value int64

	err := c.db.QueryRow(ctx, query, cityID).Scan(&year, &value)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query population for city %d: %w", cityID, err)
	}

	return &domain.Indicator[int64]{Year: year, Value: value}, nil
}

func (c *CityRepo) findIncome(ctx context.Context, cityID int64) (*domain.Indicator[float64], error) {
	const query = `
		SELECT year, average_income
		FROM income_indicators
		WHERE city_id = $1
		ORDER BY year DESC
		LIMIT 1
	`

	var year int
	var value float64

	err := c.db.QueryRow(ctx, query, cityID).Scan(&year, &value)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query income for city %d: %w", cityID, err)
	}

	return &domain.Indicator[float64]{Year: year, Value: value}, nil
}

func (c *CityRepo) findGDP(ctx context.Context, cityID int64) (*domain.Indicator[float64], error) {
	const query = `
		SELECT year, gdp
		FROM gdp_indicators
		WHERE city_id = $1
		ORDER BY year DESC
		LIMIT 1
	`

	var year int
	var value float64

	err := c.db.QueryRow(ctx, query, cityID).Scan(&year, &value)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query GDP for city %d: %w", cityID, err)
	}

	return &domain.Indicator[float64]{Year: year, Value: value}, nil
}
