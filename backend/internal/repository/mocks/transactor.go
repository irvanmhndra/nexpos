package mocks

import "context"

// Transactor runs fn directly with the caller's ctx, so unit tests keep
// matching repository expectations on that ctx. Atomicity itself is covered
// by the integration tests against PostgreSQL.
type Transactor struct{}

func (Transactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
