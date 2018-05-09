package equality

import "reflect"
import "goshua/goshua"
import "fmt"

// Two pointers are equal if they are both nil, if they point to the
// same object, or if the objects they point to are equal.

func equal_ptr_ptr(a, b interface{}) (bool, error) {
	if reflect.ValueOf(a).IsNil() {
		return reflect.ValueOf(b).IsNil(), nil
	}
	if reflect.ValueOf(b).IsNil() {
		return false, nil
	}
	// Dereference the pointers
	va := reflect.ValueOf(a).Elem()
	vb := reflect.ValueOf(b).Elem()
	if !va.IsValid() {
		return false, fmt.Errorf("dereferenced %v not valid", a)
	}
	if !vb.IsValid() {
		return false, fmt.Errorf("dereferenced %v not valid", b)
	}
	if va.UnsafeAddr() == vb.UnsafeAddr() {
		return true, nil
	}
	eq, err := goshua.Equal(va.Interface(), vb.Interface())
	if err != nil {
		return false, fmt.Errorf("%s", err)
	}
	return eq, nil
}

func init() {
	biadicDispatch[makeBiadicKey(reflect.Ptr, reflect.Ptr)] = equal_ptr_ptr
}
