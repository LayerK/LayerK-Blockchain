package validator

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

func SpawnerSupportsModule(spawner ValidationSpawner, requested common.Hash) bool {
	supported, err := spawner.WasmModuleRoots()
	if err != nil {
		log.Warn("WasmModuleRoots returned error", "err", err, "requested", requested)
		return false
	}
	return containsHash(supported, requested)
}

func containsHash(hashes []common.Hash, target common.Hash) bool {
	for _, root := range hashes {
		if root == target {
			return true
		}
	}
	return false
}
