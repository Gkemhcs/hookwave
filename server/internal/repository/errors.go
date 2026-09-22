package repository

// Hand-written, not sqlc-generated — sits next to the generated query files
// because it's the layer that actually knows about Postgres error codes.

import (
	"errors"

	"github.com/Gkemhcs/hookwave/server/internal/apierror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MapError translates a raw pgx/Postgres error into one of the package-level
// sentinel errors in apierror, so callers can check with errors.Is instead of
// knowing about Postgres error codes themselves.
func MapError(err error) error {
	var pgError *pgconn.PgError

	if errors.As(err, &pgError) {
		if pgError.Code == "23505" {
			return apierror.ErrResourceAlreadyExist
		}
		if pgError.Code == "23503" {
			return apierror.ErrParentNotExist
		}
		if pgError.Code == "23502" {
			return apierror.ErrNotNullViolation
		}
		if pgError.Code == "40P01" || pgError.Code == "40001" {
			return apierror.ErrTransactionConflict
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apierror.ErrResourceNotFound
	}
	return err
}
