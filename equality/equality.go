// Pachage equality implements a notion of equality for interface{} to
// be used by the unifier.
package equality

import "reflect"
import "goshua/goshua"

func makeBiadicKey(kind1, kind2 reflect.Kind) uint16 {
	return uint16((kind1 << 8) | kind2)
}

var biadicDispatch = make(map[uint16]func(interface{}, interface{}) (bool, error))

func equal(a, b interface{}) (bool, error) {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)
	f, ok := biadicDispatch[makeBiadicKey(va.Kind(), vb.Kind())]
	if !ok {
		return false, goshua.NewEqualError(a, b)
	}
	return f(a, b)
}

func init() {
	goshua.Equal = equal
}
