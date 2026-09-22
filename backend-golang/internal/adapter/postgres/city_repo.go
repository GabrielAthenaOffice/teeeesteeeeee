package postgres

import (
	"context"
	"fmt"

	"github.com/CunhazadanoDale/trads-market-test/internal/core/domain"
	"github.com/CunhazadanoDale/trads-market-test/internal/core/ports/out"
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
