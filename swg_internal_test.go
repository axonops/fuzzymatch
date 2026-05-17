// Copyright 2026 AxonOps Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// swg_internal_test.go covers the unexported defence-in-depth helpers
// for Smith-Waterman-Gotoh (NewSWGParams' self-test) that the public
// API surface cannot exercise. Living in package fuzzymatch lets these
// tests reach swgPanicIfInvalid directly with deliberately-invalid
// params.

package fuzzymatch

import (
	"errors"
	"math"
	"testing"
)

// TestSwgPanicIfInvalid_PanicsOnViolation covers the panic body in
// swgPanicIfInvalid (the testable internal helper for NewSWGParams).
// Reaching this branch via the public API is impossible because the
// locked default constants in NewSWGParams always validate cleanly —
// the panic only fires under build-time tampering (e.g. -ldflags
// injection). This test exercises the panic path by passing
// deliberately-invalid SWGParams to the unexported helper directly.
func TestSwgPanicIfInvalid_PanicsOnViolation(t *testing.T) {
	cases := []struct {
		name string
		bad  SWGParams
	}{
		{"Match=NaN", SWGParams{Match: math.NaN(), Mismatch: -1, GapOpen: -1, GapExtend: -0.5}},
		{"Mismatch>0", SWGParams{Match: 1, Mismatch: 1, GapOpen: -1, GapExtend: -0.5}},
		{"GapOpen=+Inf", SWGParams{Match: 1, Mismatch: -1, GapOpen: math.Inf(1), GapExtend: -0.5}},
		{"GapExtend=NaN", SWGParams{Match: 1, Mismatch: -1, GapOpen: -1, GapExtend: math.NaN()}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("swgPanicIfInvalid(%+v): expected panic, got none", c.bad)
				}
				err, ok := r.(error)
				if !ok {
					t.Fatalf("panic value is not an error: %T (%v)", r, r)
				}
				// The wrapped error must satisfy errors.Is(_, ErrInternalInvariantViolated)
				// per the documented sentinel chaining. NewSWGParams uses
				// fmt.Errorf with two %w verbs, producing a multi-error
				// (Unwrap() []error) chain — errors.Is handles both single-
				// and multi-Unwrap walks.
				if !errors.Is(err, ErrInternalInvariantViolated) {
					t.Errorf("swgPanicIfInvalid panic error chain does not contain ErrInternalInvariantViolated: %v", err)
				}
			}()
			swgPanicIfInvalid(c.bad)
			t.Fatalf("swgPanicIfInvalid(%+v): returned normally (should have panicked)", c.bad)
		})
	}
}
