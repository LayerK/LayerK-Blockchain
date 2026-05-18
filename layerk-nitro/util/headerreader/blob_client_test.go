// Copyright 2024, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro/blob/master/LICENSE.md

package headerreader

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/r3labs/diff/v3"

	"github.com/ethereum/go-ethereum/common"

	"github.com/offchainlabs/nitro/util/testhelpers"
)

func TestSaveBlobsToDisk(t *testing.T) {
	response := []blobResponseItem{{
		BlockRoot:       "a",
		Index:           0,
		Slot:            5,
		BlockParentRoot: "a0",
		ProposerIndex:   9,
		Blob:            []byte{1},
		KzgCommitment:   []byte{1},
		KzgProof:        []byte{1},
	}, {
		BlockRoot:       "a",
		Index:           1,
		Slot:            5,
		BlockParentRoot: "a0",
		ProposerIndex:   10,
		Blob:            []byte{2},
		KzgCommitment:   []byte{2},
		KzgProof:        []byte{2},
	}}
	testDir := t.TempDir()
	rawData, err := json.Marshal(response)
	Require(t, err)
	err = saveBlobDataToDisk(rawData, 5, testDir)
	Require(t, err)

	filePath := path.Join(testDir, "5")
	file, err := os.Open(filePath)
	Require(t, err)
	defer file.Close()

	data, err := io.ReadAll(file)
	Require(t, err)
	var full fullResult[[]blobResponseItem]
	err = json.Unmarshal(data, &full)
	Require(t, err)
	if !reflect.DeepEqual(full.Data, response) {
		changelog, err := diff.Diff(full.Data, response)
		Require(t, err)
		Fail(t, "blob data saved to disk does not match actual blob data", changelog)
	}
}

func TestNewBlobClientUsesSecondaryBeaconURL(t *testing.T) {
	client, err := NewBlobClient(BlobClientConfig{
		BeaconUrl:          "https://primary.example",
		SecondaryBeaconUrl: "https://secondary.example/beacon",
	}, nil)
	Require(t, err)

	if client.secondaryBeaconUrl == nil {
		Fail(t, "secondary beacon URL was not configured")
	}
	if got, want := client.secondaryBeaconUrl.String(), "https://secondary.example/beacon"; got != want {
		Fail(t, "secondary beacon URL mismatch", "got", got, "want", want)
	}
}

func TestBeaconRequestCapsAndClosesPrimaryErrorBodyBeforeSecondaryFallback(t *testing.T) {
	client, err := NewBlobClient(BlobClientConfig{
		BeaconUrl:          "https://primary.example",
		SecondaryBeaconUrl: "https://secondary.example",
	}, nil)
	Require(t, err)

	primaryBody := &closeTrackingBody{reader: strings.NewReader(strings.Repeat("x", maxBeaconErrorBodyBytes*2))}
	client.httpClient.Store(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Host {
			case "primary.example":
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Status:     "500 Internal Server Error",
					Body:       primaryBody,
					Header:     make(http.Header),
					Request:    req,
				}, nil
			case "secondary.example":
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       io.NopCloser(strings.NewReader(`{"data":{"ok":true}}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			default:
				Fail(t, "unexpected request host", req.URL.Host)
				return nil, nil
			}
		}),
	})

	data, err := beaconRequest[json.RawMessage](client, context.Background(), "/eth/v1/config/spec")
	Require(t, err)
	if got, want := string(data), `{"ok":true}`; got != want {
		Fail(t, "unexpected response data", "got", got, "want", want)
	}
	if !primaryBody.closed {
		Fail(t, "primary error response body was not closed before fallback")
	}
	if primaryBody.readBytes != maxBeaconErrorBodyBytes {
		Fail(t, "primary error response body read was not capped", "got", primaryBody.readBytes, "want", maxBeaconErrorBodyBytes)
	}
}

func TestBeaconRequestCapsSuccessfulResponseBody(t *testing.T) {
	client, err := NewBlobClient(BlobClientConfig{
		BeaconUrl: "https://beacon.example",
	}, nil)
	Require(t, err)

	client.httpClient.Store(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader(strings.Repeat("x", maxBeaconResponseBodyBytes+1))),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
	})

	_, err = beaconRequest[json.RawMessage](client, context.Background(), "/eth/v1/config/spec")
	if err == nil {
		Fail(t, "expected oversized beacon response to fail")
	}
}

func TestVersionedHashOutputIndexesKeepsFirstDuplicate(t *testing.T) {
	var first common.Hash
	var second common.Hash
	first[0] = 1
	second[0] = 2

	indexes := versionedHashOutputIndexes([]common.Hash{first, second, first})

	if got, want := indexes[first], 0; got != want {
		Fail(t, "unexpected first hash index", "got", got, "want", want)
	}
	if got, want := indexes[second], 1; got != want {
		Fail(t, "unexpected second hash index", "got", got, "want", want)
	}
	if got, want := len(indexes), 2; got != want {
		Fail(t, "unexpected index count", "got", got, "want", want)
	}
}

type closeTrackingBody struct {
	reader    *strings.Reader
	closed    bool
	readBytes int
}

func (b *closeTrackingBody) Read(p []byte) (int, error) {
	n, err := b.reader.Read(p)
	b.readBytes += n
	return n, err
}

func (b *closeTrackingBody) Close() error {
	b.closed = true
	return nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func Require(t *testing.T, err error, printables ...interface{}) {
	t.Helper()
	testhelpers.RequireImpl(t, err, printables...)
}

func Fail(t *testing.T, printables ...interface{}) {
	t.Helper()
	testhelpers.FailImpl(t, printables...)
}
