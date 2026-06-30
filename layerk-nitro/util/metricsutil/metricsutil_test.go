// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro/blob/master/LICENSE.md

package metricsutil

import "testing"

func TestCanonicalizeMetricName(t *testing.T) {
	tests := []struct {
		name   string
		metric string
		want   string
	}{
		{
			name:   "leaves valid chars",
			metric: "metric_name:total42",
			want:   "metric_name:total42",
		},
		{
			name:   "replaces invalid runs",
			metric: "https://example.com:8545/rpc?chain=layerk",
			want:   "https:_example_com:8545_rpc_chain_layerk",
		},
		{
			name:   "keeps separator per invalid run",
			metric: "alpha-beta/gamma",
			want:   "alpha_beta_gamma",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := CanonicalizeMetricName(test.metric); got != test.want {
				t.Fatalf("CanonicalizeMetricName(%q) = %q, want %q", test.metric, got, test.want)
			}
		})
	}
}
