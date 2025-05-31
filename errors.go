package fixedPoint

import "errors"

var (
	ErrOverflow    = errors.New("fixedPoint: overflow")
	ErrNegOverflow = errors.New("fixedPoint: negative overflow")
	ErrUnderflow   = errors.New("fixedPoint: underflow")
	ErrDivByZero   = errors.New("fixedPoint: division by zero")
	ErrDomain      = errors.New("fixedPoint: input out of domain")
)
