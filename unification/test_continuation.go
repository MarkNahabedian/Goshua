package unification

import "testing"
import "goshua/goshua"

type testContinuation struct {
	continued bool
	bindings  goshua.Bindings
	t         *testing.T
}

func (tc *testContinuation) Continuation(b goshua.Bindings) {
	tc.t.Logf("Continuation called")
	tc.continued = true
	tc.bindings = b
}

func (tc *testContinuation) WasContinued() bool {
	return tc.continued
}

func (tc *testContinuation) Bindings() goshua.Bindings {
	if !tc.WasContinued() {
		panic("Tryingto get Bindings from a testContinuation that was never continued")
	}
	return tc.bindings
}

func MakeTestContinuation(t *testing.T) *testContinuation {
	tc := testContinuation{
		continued: false,
		t:         t,
	}
	return &tc
}
