// Copyright 2022, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro/blob/master/LICENSE.md

package gethexec

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
)

func TestClassicOutboxProofNodesKeepLeafToRootOrder(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	batchNum := big.NewInt(7)
	root := classicOutboxTestHash(1)
	leftNode := classicOutboxTestHash(2)
	rightLeaf := classicOutboxTestHash(3)
	leftLeaf := classicOutboxTestHash(4)
	targetLeaf := classicOutboxTestHash(5)
	targetData := []byte("target")

	putClassicOutboxBatch(t, db, batchNum, 3, root)
	putClassicOutboxNode(t, db, root, leftNode, rightLeaf)
	putClassicOutboxNode(t, db, leftNode, leftLeaf, targetLeaf)
	putClassicOutboxLeaf(t, db, targetLeaf, targetData)

	msg, err := NewClassicOutboxRetriever(db).GetMsg(batchNum, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(msg.Data, targetData) {
		t.Fatalf("unexpected data %q", msg.Data)
	}
	if msg.PathInt.Cmp(big.NewInt(2)) != 0 {
		t.Fatalf("unexpected path int %s", msg.PathInt)
	}
	if len(msg.ProofNodes) != 2 {
		t.Fatalf("expected 2 proof nodes, got %d", len(msg.ProofNodes))
	}
	if !bytes.Equal(msg.ProofNodes[0][:], leftLeaf[:]) {
		t.Fatalf("unexpected leaf-level proof node %x", msg.ProofNodes[0])
	}
	if !bytes.Equal(msg.ProofNodes[1][:], rightLeaf[:]) {
		t.Fatalf("unexpected root-level proof node %x", msg.ProofNodes[1])
	}
}

func TestClassicOutboxRejectsIndexAtMerkleSize(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	batchNum := big.NewInt(8)
	root := classicOutboxTestHash(6)

	putClassicOutboxBatch(t, db, batchNum, 1, root)
	putClassicOutboxLeaf(t, db, root, []byte("leaf"))

	_, err := NewClassicOutboxRetriever(db).GetMsg(batchNum, 1)
	if err == nil {
		t.Fatal("expected out-of-range error")
	}
	if !strings.Contains(err.Error(), "only has 1 indexes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

type classicOutboxTestPutter interface {
	Put(key []byte, value []byte) error
}

func classicOutboxTestHash(value byte) common.Hash {
	var hash common.Hash
	hash[31] = value
	return hash
}

func putClassicOutboxBatch(t *testing.T, db classicOutboxTestPutter, batchNum *big.Int, size uint64, root common.Hash) {
	t.Helper()
	header := make([]byte, 40)
	binary.BigEndian.PutUint64(header[:8], size)
	copy(header[8:40], root[:])
	if err := db.Put(msgBatchKey(batchNum), header); err != nil {
		t.Fatal(err)
	}
}

func putClassicOutboxNode(t *testing.T, db classicOutboxTestPutter, root, left, right common.Hash) {
	t.Helper()
	node := make([]byte, 64)
	copy(node[:32], left[:])
	copy(node[32:64], right[:])
	if err := db.Put(root[:], node); err != nil {
		t.Fatal(err)
	}
}

func putClassicOutboxLeaf(t *testing.T, db classicOutboxTestPutter, root common.Hash, data []byte) {
	t.Helper()
	if err := db.Put(root[:], data); err != nil {
		t.Fatal(err)
	}
}
