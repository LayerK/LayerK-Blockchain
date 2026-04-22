// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro-contracts/blob/main/LICENSE
// SPDX-License-Identifier: BUSL-1.1

pragma solidity ^0.8.4;

import {MerkleProofTooLong} from "./Error.sol";

library MerkleLib {
    function generateRoot(
        bytes32[] memory _hashes
    ) internal pure returns (bytes32) {
        bytes32[] memory prevLayer = _hashes;
        while (prevLayer.length > 1) {
            uint256 prevLength = prevLayer.length;
            uint256 nextLength = (prevLength + 1) / 2;
            bytes32[] memory nextLayer = new bytes32[](nextLength);
            for (uint256 i = 0; i < nextLength;) {
                uint256 left = 2 * i;
                if (left + 1 < prevLength) {
                    nextLayer[i] = keccak256(abi.encodePacked(prevLayer[left], prevLayer[left + 1]));
                } else {
                    nextLayer[i] = prevLayer[left];
                }
                unchecked {
                    ++i;
                }
            }
            prevLayer = nextLayer;
        }
        return prevLayer[0];
    }

    function calculateRoot(
        bytes32[] memory nodes,
        uint256 route,
        bytes32 item
    ) internal pure returns (bytes32) {
        return _calculateRoot(nodes.length, route, item, nodes);
    }

    function calculateRoot(
        bytes32[] calldata nodes,
        uint256 route,
        bytes32 item
    ) internal pure returns (bytes32) {
        return _calculateRoot(nodes.length, route, item, nodes);
    }

    function _calculateRoot(
        uint256 proofItems,
        uint256 route,
        bytes32 item,
        bytes32[] memory nodes
    ) private pure returns (bytes32) {
        if (proofItems > 256) revert MerkleProofTooLong(proofItems, 256);
        bytes32 h = item;
        for (uint256 i = 0; i < proofItems;) {
            bytes32 node = nodes[i];
            if ((route & (1 << i)) == 0) {
                assembly {
                    mstore(0x00, h)
                    mstore(0x20, node)
                    h := keccak256(0x00, 0x40)
                }
            } else {
                assembly {
                    mstore(0x00, node)
                    mstore(0x20, h)
                    h := keccak256(0x00, 0x40)
                }
            }
            unchecked {
                ++i;
            }
        }
        return h;
    }

    function _calculateRoot(
        uint256 proofItems,
        uint256 route,
        bytes32 item,
        bytes32[] calldata nodes
    ) private pure returns (bytes32) {
        if (proofItems > 256) revert MerkleProofTooLong(proofItems, 256);
        bytes32 h = item;
        for (uint256 i = 0; i < proofItems;) {
            bytes32 node = nodes[i];
            if ((route & (1 << i)) == 0) {
                assembly {
                    mstore(0x00, h)
                    mstore(0x20, node)
                    h := keccak256(0x00, 0x40)
                }
            } else {
                assembly {
                    mstore(0x00, node)
                    mstore(0x20, h)
                    h := keccak256(0x00, 0x40)
                }
            }
            unchecked {
                ++i;
            }
        }
        return h;
    }
}
