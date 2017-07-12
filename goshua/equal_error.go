package goshua

import "fmt"

// equalError is the type of error returned as the second value of Equal.
type equalError struct {
	arg1 interface{}
	arg2 interface{}
}

func (e *equalError) Error() string {
	return fmt.Sprintf("goshual.Equal doesn't know how to compare %T with %t",
		e.arg1, e.arg2)
}

func (e *equalError) Arg1() interface{} {
	return e.arg1
}

func (e *equalError) Arg2() interface{} {
	return e.arg2
}

// NewEqualError creates an error to be returned as the second value of
// goshua.Equal.
func NewEqualError(arg1, arg2 interface{}) *equalError {
	return &equalError{
		arg1: arg1,
		arg2: arg2,
	}
}
