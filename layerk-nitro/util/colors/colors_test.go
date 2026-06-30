// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro/blob/master/LICENSE.md

package colors

import "testing"

func TestUncolor(t *testing.T) {
	input := Red + "hello\t" + Clear + "\n" + Blue + "world" + Clear
	want := "hello world"

	if got := Uncolor(input); got != want {
		t.Fatalf("Uncolor(%q) = %q, want %q", input, got, want)
	}
}
