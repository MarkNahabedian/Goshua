// Pachage equality implements a notion of equality for interface{} to
// be used by the unifier.
package equality

import "reflect"
import "goshua/goshua"

func makeBiadicKey(kind1, kind2 reflect.Kind) uint16 {
	return uint16((kind1 << 8) | kind2)
}

var biadicDispatch = make(map[uint16]func(interface{}, interface{}) bool)

func equal(a, b interface{}) (bool, error) {
	f, ok := biadicDispatch[makeBiadicKey(
		reflect.ValueOf(a).Kind(),
		reflect.ValueOf(b).Kind())]
	if !ok {
		return false, goshua.NewEqualError(a, b)
	}
	return f(a, b), nil
}

func init() {
	goshua.Equal = equal
}
