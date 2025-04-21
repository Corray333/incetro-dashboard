package errs

import "errors"

var (
	ErrInvalidTxTypeInCtx = errors.New("invalid transaction in context")
)
