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

// kind.go declares the Kind enum that classifies a scan.Warning's pair
// scope (within-group vs cross-group). The shape mirrors the root
// fuzzymatch.WarnKind type verbatim — int with iota+1 constants, a
// switch-based String() that returns the CamelCase suffix label, and
// an allocating fmt.Sprintf default branch for out-of-range values.
//
// Implementation discipline (mirrors warn_kind.go):
//
//   - No init() — String() is a switch returning compile-time string
//     constants. No map iteration on output paths (per
//     .claude/skills/determinism-standards).
//   - Constants start at iota + 1 so the zero value of Kind is
//     "unspecified" — defensive against accidental zero-initialisation
//     emitting a misleading WithinGroup.

package scan

import "fmt"

// Kind classifies a scan.Warning's pair scope: within-group (both
// items share the same Group) or cross-group (items have different
// Group values). Kind is a plain int (not int32 / int64 / struct) so
// it is trivially comparable and serialisable. The zero value is
// reserved as "unspecified" — every documented Kind starts at 1
// (iota + 1).
//
// SPEC OVERRIDE (Phase 9): docs/requirements.md §12.1 originally
// specified this type as "WarningKind". The implementation uses Kind
// because the package-scoped form (scan.KindWithinGroup,
// scan.KindAcrossGroups) is unambiguous at the call site and avoids
// accidental symmetry with the root package's WarnKind type (which
// classifies Validate diagnostics, a different domain). The spec
// deviation is documented in 09-CONTEXT.md §1 D-02; the
// api-ergonomics-reviewer signed off on this override in plan 09-01's
// PR. The spec was amended in lockstep with this declaration.
//
// Values are stable across patch releases. The integer values
// themselves are part of the v1.x contract: consumers may persist
// them, compare them, and rely on KindWithinGroup evaluating to 1.
// Future additions append to the END of the const block — existing
// Kind values never shift.
//
// Use String() to obtain the canonical CamelCase label
// ("WithinGroup", "AcrossGroups").
type Kind int

// The two Kind constants documented in docs/requirements.md §12.1 (as
// amended by Phase 9 — see 09-CONTEXT.md §1 D-02). Iota starts at 1 so
// the zero value remains "unset" — useful for detecting accidentally
// zero-initialised Warning values in consumer code.
const (
	// KindWithinGroup signals a similar-name pair where both items
	// have the same Group value. Emitted by the within-group pass.
	KindWithinGroup Kind = iota + 1

	// KindAcrossGroups signals a similar-name pair where the items
	// have different Group values. Emitted by the cross-group pass
	// when Config.CompareAcrossGroups is true.
	KindAcrossGroups
)

// String returns the canonical CamelCase label for k. For
// KindWithinGroup the label is "WithinGroup" — the constant prefix is
// dropped to match the AlgoID.String() and WarnKind.String() naming
// convention locked in docs/requirements.md §6 and Phase 8.5 Q6b.
//
// For an unknown Kind value (the zero value, or any future value
// declared after this method is compiled), String returns the fallback
// "Kind(N)" via fmt.Sprintf — intentionally allocating because the
// path is for diagnostic and error output, not a hot dispatch path.
//
// String never allocates on the in-range path: every case returns a
// compile-time string constant.
func (k Kind) String() string {
	switch k {
	case KindWithinGroup:
		return "WithinGroup"
	case KindAcrossGroups:
		return "AcrossGroups"
	default:
		return fmt.Sprintf("Kind(%d)", int(k))
	}
}
