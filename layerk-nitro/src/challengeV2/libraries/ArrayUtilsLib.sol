// Copyright 2023, Offchain Labs, Inc.
// For license information, see https://github.com/offchainlabs/bold/blob/main/LICENSE
// SPDX-License-Identifier: BUSL-1.1
//
pragma solidity ^0.8.17;

/// @title  Array utils library
/// @notice Utils for working with bytes32 arrays
library ArrayUtilsLib {
    /// @notice Append an item to the end of an array
    /// @param arr      The array to append to
    /// @param newItem  The item to append
    function append(
        bytes32[] memory arr,
        bytes32 newItem
    ) internal pure returns (bytes32[] memory) {
        uint256 arrLength = arr.length;
        bytes32[] memory clone = new bytes32[](arrLength + 1);
        for (uint256 i = 0; i < arrLength;) {
            clone[i] = arr[i];
            unchecked {
                ++i;
            }
        }
        clone[arrLength] = newItem;
        return clone;
    }

    /// @notice Get a slice of an existing array
    /// @dev    End index exlusive so slice(arr, 0, 5) gets the first 5 elements of arr
    /// @param arr          Array to slice
    /// @param startIndex   The start index of the slice in the original array - inclusive
    /// @param endIndex     The end index of the slice in the original array - exlusive
    function slice(
        bytes32[] memory arr,
        uint256 startIndex,
        uint256 endIndex
    ) internal pure returns (bytes32[] memory) {
        require(startIndex < endIndex, "Start not less than end");
        require(endIndex <= arr.length, "End not less or equal than length");

        bytes32[] memory newArr = new bytes32[](endIndex - startIndex);
        for (uint256 i = startIndex; i < endIndex;) {
            newArr[i - startIndex] = arr[i];
            unchecked {
                ++i;
            }
        }
        return newArr;
    }

    /// @notice Concatenated to arrays
    /// @param arr1 First array
    /// @param arr2 Second array
    function concat(
        bytes32[] memory arr1,
        bytes32[] memory arr2
    ) internal pure returns (bytes32[] memory) {
        uint256 arr1Length = arr1.length;
        uint256 arr2Length = arr2.length;
        bytes32[] memory full = new bytes32[](arr1Length + arr2Length);
        for (uint256 i = 0; i < arr1Length;) {
            full[i] = arr1[i];
            unchecked {
                ++i;
            }
        }
        for (uint256 i = 0; i < arr2Length;) {
            full[arr1Length + i] = arr2[i];
            unchecked {
                ++i;
            }
        }
        return full;
    }
}
