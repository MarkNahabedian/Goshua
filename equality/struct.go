package equality

import "reflect"
import "goshua/goshua"

// Two structs are equal if they are the same type and have the same content.

func equal_struct_struct(a, b interface{}) (bool, error) {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)
	if va.Type() != vb.Type() {
		return false, nil
	}
	for i := 0; i < va.NumField(); i++ {
		eq, err := goshua.Equal(
			va.Field(i).Interface(),
			vb.Field(i).Interface())
		if err != nil {
			return false, nil
		}
		if !eq {
			return false, nil
		}
	}
	return true, nil
}

func init() {
	biadicDispatch[makeBiadicKey(reflect.Struct, reflect.Struct)] = equal_struct_struct
}
