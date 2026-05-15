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

// phonetic_codes_golden_test.go pins the encoded-code output of Phase 7
// phonetic algorithms byte-for-byte across the 5-platform CI matrix. It
// uses a separate golden file (testdata/golden/phonetic-codes.json) from
// the float-score golden file (testdata/golden/algorithms.json) per
// CONTEXT.md §7 LOCKED — the two schemas are structurally different (string
// codes vs float64 scores) and must not be merged.
//
// Soundex loader is ACTIVE in plan 07-01; DoubleMetaphone / NYSIIS / MRA
// loaders are stubbed with t.Skip for plans 07-02..07-04.
//
// Schema of testdata/golden/phonetic-codes.json:
//
//	{
//	  "_metadata": {
//	    "purpose": "Cross-platform byte-stable phonetic code determinism gate",
//	    "regenerated_at": "<ISO>"
//	  },
//	  "entries": [
//	    {"algorithm": "Soundex", "input": "Robert", "code": "R163"},
//	    {"algorithm": "DoubleMetaphone", "input": "Schmidt",
//	     "primary": "XMT", "secondary": "SMT"},
//	    ...
//	  ]
//	}
//
// String equality (not float-bit equality) is the assertion: the encoded
// code must match exactly on every CI platform (linux/amd64, linux/arm64,
// darwin/amd64, darwin/arm64, windows/amd64).

package fuzzymatch_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// phoneticCodeGoldenEntry is one (algorithm, input, expected-code) case in
// testdata/golden/phonetic-codes.json.
type phoneticCodeGoldenEntry struct {
	Algorithm string `json:"algorithm"`
	Input     string `json:"input"`
	Code      string `json:"code,omitempty"`      // Soundex / NYSIIS / MRA
	Primary   string `json:"primary,omitempty"`   // DoubleMetaphone primary key
	Secondary string `json:"secondary,omitempty"` // DoubleMetaphone secondary key
}

// phoneticCodeGoldenMetadata is the _metadata block of phonetic-codes.json.
type phoneticCodeGoldenMetadata struct {
	Purpose       string `json:"purpose"`
	RegeneratedAt string `json:"regenerated_at"`
}

// phoneticCodeGoldenFile is the top-level shape of phonetic-codes.json.
type phoneticCodeGoldenFile struct {
	Metadata phoneticCodeGoldenMetadata `json:"_metadata"`
	Entries  []phoneticCodeGoldenEntry  `json:"entries"`
}

// TestPhoneticCodesGolden asserts byte-stable encoded-code output for each
// Phase 7 phonetic algorithm across the 5-platform CI matrix.
// Soundex sub-test ASSERTS (plan 07-01); DM/NYSIIS/MRA stub t.Skip for
// plans 07-02..07-04.
func TestPhoneticCodesGolden(t *testing.T) {
	path := filepath.Join("testdata", "golden", "phonetic-codes.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("TestPhoneticCodesGolden: read %s: %v", path, err)
	}
	var gf phoneticCodeGoldenFile
	if err := json.Unmarshal(raw, &gf); err != nil {
		t.Fatalf("TestPhoneticCodesGolden: parse %s: %v", path, err)
	}
	if len(gf.Entries) == 0 {
		t.Fatalf("TestPhoneticCodesGolden: empty golden file")
	}

	t.Run("Soundex", func(t *testing.T) {
		n := 0
		for _, e := range gf.Entries {
			e := e
			if e.Algorithm != "Soundex" {
				continue
			}
			n++
			t.Run(e.Input, func(t *testing.T) {
				got := fuzzymatch.SoundexCode(e.Input)
				if got != e.Code {
					t.Errorf("SoundexCode(%q) = %q; golden wants %q (byte-stable cross-platform gate)",
						e.Input, got, e.Code)
				}
			})
		}
		if n == 0 {
			t.Fatal("no Soundex entries in phonetic-codes.json golden file")
		}
	})

	t.Run("DoubleMetaphone", func(t *testing.T) {
		n := 0
		for _, e := range gf.Entries {
			e := e
			if e.Algorithm != "DoubleMetaphone" {
				continue
			}
			n++
			t.Run(e.Input, func(t *testing.T) {
				gotP, gotS := fuzzymatch.DoubleMetaphoneKeys(e.Input)
				if gotP != e.Primary {
					t.Errorf("DoubleMetaphoneKeys(%q).primary = %q; golden wants %q (byte-stable cross-platform gate)",
						e.Input, gotP, e.Primary)
				}
				if gotS != e.Secondary {
					t.Errorf("DoubleMetaphoneKeys(%q).secondary = %q; golden wants %q (byte-stable cross-platform gate)",
						e.Input, gotS, e.Secondary)
				}
			})
		}
		if n == 0 {
			t.Fatal("no DoubleMetaphone entries in phonetic-codes.json golden file")
		}
	})

	t.Run("NYSIIS", func(t *testing.T) {
		t.Skip("enabled by plan 07-03")
	})

	t.Run("MRA", func(t *testing.T) {
		t.Skip("enabled by plan 07-04")
	})
}
