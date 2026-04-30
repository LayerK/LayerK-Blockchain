// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro-contracts/blob/main/LICENSE
// SPDX-License-Identifier: BUSL-1.1

pragma solidity ^0.8.0;

library CallerChecker {
    /// @notice Returns true when the immediate caller is both the transaction origin and has no
    ///         deployed code (i.e. is a plain EOA calling at the top level of a tx).
    /// @dev    This is used to guard functions that should only be reachable from a direct EOA tx
    ///         and to ensure gas-refund accounting can trust that calldata came from tx.input.
    // solhint-disable-next-line avoid-tx-origin
    function isCallerCodelessOrigin() internal view returns (bool) {
        // solhint-disable-next-line avoid-tx-origin
        return msg.sender == tx.origin && msg.sender.code.length == 0;
    }
}
