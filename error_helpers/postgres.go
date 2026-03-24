package error_helpers

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func DecodePgError(err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return errors.New(pgError.Message)
	}
	return err
}
