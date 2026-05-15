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

// Package main demonstrates the four Phase 7 phonetic encoding functions
// (SoundexCode, DoubleMetaphoneKeys, NYSIISCode, MRACode + MRACompare)
// on a curated set of English-language surname samples. This program
// complements the score-focused identifier-similarity example by showing
// the encoded-key surfaces standalone — the raw phonetic codes that each
// algorithm assigns to an input name.
//
// The curated name set is drawn from the Phase 7 reference vectors and
// covers a range of surname origins: Anglo-Saxon (Robert, Brown), Germanic
// (Schmidt, Boern), Romance/Spanish (Pacheco), and phonetically similar
// pairs (Byrne/Boern, Catherine/Katherine, Brown/Browne) that exercise the
// cross-algorithm matching behaviour.
//
// Run with:
//
//	go run ./examples/phonetic-keys/
package main

import (
	"fmt"
	"strings"

	"github.com/axonops/fuzzymatch"
)

// names is the curated list of English-language surnames used to demonstrate
// the four Phase 7 phonetic encoding functions. The set is drawn from the
// Phase 7 literature reference vectors (Knuth TAOCP §6.4 for Soundex/NYSIIS,
// Philips 2000 for Double Metaphone, NBS Tech Note 943 for MRA) and covers
// phonetically similar pairs, variant-gate names (Tymczak → T522), and
// multi-language-origin surnames (Pacheco for the Romance branch).
var names = []string{
	"Robert", "Rupert", "Tymczak", "Schmidt", "Smith",
	"Catherine", "Katherine", "Brown", "Browne", "Pacheco", "Byrne", "Boern",
}

func main() {
	// Header: Name | Soundex | DM-primary | DM-secondary | NYSIIS | MRA
	// The last column uses %s (not %-8s) to avoid trailing spaces that would
	// differ across platforms with varying trailing-whitespace handling.
	fmt.Printf("%-12s %-8s %-8s %-8s %-8s %s\n",
		"Name", "Soundex", "DM-pri", "DM-sec", "NYSIIS", "MRA")
	fmt.Println(strings.Repeat("-", 52))

	// Data rows: one per name, showing encoded keys side-by-side.
	for _, n := range names {
		primary, secondary := fuzzymatch.DoubleMetaphoneKeys(n)
		fmt.Printf("%-12s %-8s %-8s %-8s %-8s %s\n",
			n,
			fuzzymatch.SoundexCode(n),
			primary, secondary,
			fuzzymatch.NYSIISCode(n),
			fuzzymatch.MRACode(n))
	}

	// MRACompare examples: demonstrates the (bool, int) return surface from
	// NBS Tech Note 943 step 5/6. The int is the raw 0-6 NBS similarity
	// counter; the bool is the threshold-rule match decision.
	fmt.Println()
	fmt.Println("MRACompare examples:")
	pairs := [][2]string{
		{"Byrne", "Boern"},
		{"Smith", "Smyth"},
		{"Catherine", "Katherine"},
		{"Ad", "ZachariahMontgomery"},
	}
	for _, p := range pairs {
		matched, sim := fuzzymatch.MRACompare(p[0], p[1])
		fmt.Printf("  %s vs %s: matched=%v sim=%d\n", p[0], p[1], matched, sim)
	}
}
