package out

import "context"

type AgeRepository interface {
	Upsert(
		ctx context.Context,
		ibgeCode int64,
		year int,
		ageGroup string,
		population int64,
	) error
}
