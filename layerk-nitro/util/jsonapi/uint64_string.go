// Copyright 2024, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro/blob/master/LICENSE.md

package jsonapi

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// Uint64String is a uint64 that JSON marshals and unmarshals as string in decimal
type Uint64String uint64

func (u *Uint64String) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		return nil
	}

	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}

	*u = Uint64String(value)
	return nil
}

func (u Uint64String) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 0, 24)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, uint64(u), 10)
	buf = append(buf, '"')
	return buf, nil
}
