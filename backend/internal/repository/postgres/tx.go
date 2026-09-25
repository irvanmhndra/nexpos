package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

// querier is the query surface shared by *sqlx.DB and *sqlx.Tx.
type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
}

// conn returns the transaction carried by ctx (see TxManager.WithinTx), or db
// when there is none. Every repository query goes through it, so a method
// joins the caller's transaction without taking a second pool connection.
func conn(ctx context.Context, db *sqlx.DB) querier {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return db
}

// errNoTx guards locking reads: outside a transaction a row lock is released
// as soon as the statement ends, so it would silently protect nothing.
var errNoTx = errors.New("locking read requires a transaction (use Transactor.WithinTx)")

func requireTx(ctx context.Context) error {
	if _, ok := ctx.Value(txKey{}).(*sqlx.Tx); !ok {
		return errNoTx
	}
	return nil
}

// TxManager implements repository.Transactor.
type TxManager struct {
	db *sqlx.DB
}

func NewTxManager(db *sqlx.DB) *TxManager {
	return &TxManager{db: db}
}

// WithinTx runs fn in one transaction, committing when fn returns nil and
// rolling back otherwise (including on panic). A nested call joins the
// outer transaction.
func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if _, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return fn(ctx)
	}
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return err
	}
	return tx.Commit()
}
