package perr

import (
	"github.com/rs/xid"
)

func reference() string {
	return "fperr_" + xid.New().String()
}

func IsPerr(err error) bool {
	_, ok := err.(ErrorModel)
	return ok
}
