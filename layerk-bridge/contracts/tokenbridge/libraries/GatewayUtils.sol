// SPDX-License-Identifier: Apache-2.0

/*
 * Copyright 2020, Offchain Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

pragma solidity ^0.8.0;

/// @notice Collection of gateway utility functions
library GatewayUtils {
    /// @notice Compute a create2 salt from a gateway and token address
    /// @param counterpartGateway The counterpart gateway address
    /// @param l1ERC20 The L1 token address
    /// @return salt The computed salt value
    function computeSalt(address counterpartGateway, address l1ERC20)
        internal
        pure
        returns (bytes32 salt)
    {
        salt = keccak256(
            abi.encode(counterpartGateway, keccak256(abi.encode(l1ERC20)))
        );
    }
}
