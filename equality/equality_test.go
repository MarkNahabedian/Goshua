package equality

import "testing"
import "goshua/goshua"

func test(t *testing.T, equal bool, a, b interface{}) {
	eq, err := goshua.Equal(a, b)
	if err != nil {
		t.Errorf("%s", err.Error())
		return
	}
	if equal {
		if !eq {
			t.Errorf("expected %T(%v) and %T(%v) to be equal", a, a, b, b)
		}
	} else {
		if eq {
			t.Errorf("expected %T(%v) and %T(%v) to not be equal", a, a, b, b)
		}
	}
}

func TestInt(t *testing.T) {
	test(t, true, int8(23), int8(23))
	test(t, true, int8(23), int32(23))
	test(t, false, int16(23), int16(100))
	test(t, false, int16(45), int32(23))
}

func TestUint(t *testing.T) {
	test(t, true, uint8(23), uint8(23))
	test(t, true, uint8(23), uint32(23))
	test(t, false, uint16(23), uint16(100))
	test(t, false, uint16(45), uint32(23))
}

func TestIntUint(t *testing.T) {
	test(t, true, int8(69), uint8(69))
	test(t, false, ^int8(0), ^uint8(0))
	test(t, false, ^int(0), ^uint(0))
}

func TestUintInt(t *testing.T) {
	rtest := func(equal bool, a, b interface{}) {
		test(t, equal, b, a)
	}
	rtest(true, int8(69), uint8(69))
	rtest(false, ^int8(0), ^uint8(0))
	rtest(false, ^int(0), ^uint(0))
}

func TestFloat(t *testing.T) {
	test(t, true, float32(3.14), float32(3.14))
	//  test(t, true, float32(3.14), float64(3.14))
	test(t, false, float32(3.14), float32(1.414))
}

/*
func TestComplex(t *testing.T) {
  test(t, true, complex64(3+4i), complex64(3+4i))
  test(t, true, complex64(3+4i), complex128(3+4i))
  test(t, false, complex64(4+3i), complex128(3+4i))
}
*/

func TestString(t *testing.T) {
	test(t, true, "abcdef", "abcdef")
	test(t, false, "abcdef", "ABCDEF")
}

type testStruct struct {
	A int
	B string
	C *testStruct
}

func TestStruct(t *testing.T) {
	ts1 := testStruct{
		A: 1,
		B: "foo",
	}
	ts2 := testStruct{
		A: 1,
		B: "foo",
	}
	ts3 := testStruct{
		A: 1,
		B: "bar",
	}
	test(t, true, ts1, ts1)
	test(t, true, ts1, ts2)
	test(t, false, ts1, ts3)
	test(t, true, &ts1, &ts1)
	test(t, true, &ts1, &ts2)
	test(t, false, &ts1, &ts3)

}
