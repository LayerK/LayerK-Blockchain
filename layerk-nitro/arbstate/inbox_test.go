// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro/blob/master/LICENSE.md

package arbstate

import (
	"bytes"
	"testing"
)

func TestReadSequencerPayloadWithLimitAllowsExactLimit(t *testing.T) {
	input := []byte{1, 2, 3}
	payload, err := readSequencerPayloadWithLimit(bytes.NewReader(input), len(input))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(payload, input) {
		t.Fatalf("got payload %v, wanted %v", payload, input)
	}
}

func TestReadSequencerPayloadWithLimitRejectsOversizedPayload(t *testing.T) {
	payload, err := readSequencerPayloadWithLimit(bytes.NewReader([]byte{1, 2, 3, 4}), 3)
	if err == nil {
		t.Fatal("expected oversized payload error")
	}
	if payload != nil {
		t.Fatalf("expected nil payload on error, got %v", payload)
	}
}
