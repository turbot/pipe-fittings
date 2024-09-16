package error_helpers

import (
	"errors"
	"github.com/jackc/pgconn"
)

func DecodePgError(err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return errors.New(pgError.Message)
	}
	return err
}
