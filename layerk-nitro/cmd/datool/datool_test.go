// Copyright 2026, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro/blob/master/LICENSE.md

package main

import (
	"strings"
	"testing"
)

func TestStartClientRequiresSubcommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing client subcommand",
			args: nil,
			want: "datool client requires 'rpc' or 'rest'",
		},
		{
			name: "missing rpc subcommand",
			args: []string{"rpc"},
			want: "datool client rpc requires 'store'",
		},
		{
			name: "missing rest subcommand",
			args: []string{"rest"},
			want: "datool client rest requires 'getByHash'",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requireErrContains(t, startClient(test.args), test.want)
		})
	}
}

func TestStartClientStoreShortSigningKeyReturnsError(t *testing.T) {
	requireErrContains(t, startClientStore([]string{"--signing-key", "x"}), "x")
}

func requireErrContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("unexpected error: got %q want substring %q", err, want)
	}
}
