package postgres

import (
	"context"
	"fmt"

	"github.com/CunhazadanoDale/trads-market-test/internal/core/ports/out"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ out.AgeRepository = (*AgeRepo)(nil)

type AgeRepo struct {
	db *pgxpool.Pool
}

func NewAgeRepo(db *pgxpool.Pool) *AgeRepo {
	return &AgeRepo{
		db: db,
	}
}

func (a *AgeRepo) Upsert(
	ctx context.Context,
	ibgeCode int64,
	year int,
	ageGroup string,
	population int64,
) error {
	query := `INSERT INTO age_indicators (
			city_id,
			year,
			age_group,
			population
		)
		SELECT
			id,
			$2,
			$3,
			$4
		FROM cities
		WHERE ibge_code = $1
		ON CONFLICT (city_id, year, age_group)
		DO UPDATE SET
			population = EXCLUDED.population,
			updated_at = NOW()
	`

	result, err := a.db.Exec(
		ctx,
		query,
		ibgeCode,
		year,
		ageGroup,
		population,
	)
	if err != nil {
		return fmt.Errorf(
			"upsert age %q for IBGE code %d: %w",
			ageGroup,
			ibgeCode,
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf(
			"city not found for IBGE code %d",
			ibgeCode,
		)
	}

	return nil
}
