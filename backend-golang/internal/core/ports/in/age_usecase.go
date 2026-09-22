package in

import "context"

type AgeUsecase interface {
	Import(ctx context.Context) error
}
