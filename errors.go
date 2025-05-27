package fixedPoint

import "errors"

var (
	ErrOverflow    = errors.New("fix64: overflow")
	ErrNegOverflow = errors.New("fix64: negative overflow")
	ErrUnderflow   = errors.New("fix64: underflow")
	ErrDivByZero   = errors.New("fix64: division by zero")
	ErrDomain      = errors.New("fix64: input out of domain")
)
