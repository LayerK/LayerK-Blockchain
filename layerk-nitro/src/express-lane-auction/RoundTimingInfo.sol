// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.0;

import {NegativeRoundStart} from "./Errors.sol";

/// @notice Information about the timings of auction round. All timings measured in seconds
///         Bids can be submitted to the offchain autonomous auctioneer until the auction round closes
///         after which the auctioneer can submit the two highest bids to the auction contract to resolve the auction
struct RoundTimingInfo {
    /// @notice The timestamp when round 0 starts
    ///         We allow this to be negative so that later when setting new round timing info
    ///         we can use offsets very far in the past. This combined with the maxium that we allow
    ///         on round duration ensures that we can always set a new round duration within the possible range
    int64 offsetTimestamp;
    /// @notice The total duration (in seconds) of the round. Always less than 86400 and greater than 0
    uint64 roundDurationSeconds;
    /// @notice The number of seconds before the end of the round that the auction round closes. Cannot be 0
    uint64 auctionClosingSeconds;
    /// @notice A reserve setter account has the rights to set a reserve for a round,
    ///         however they cannot do this within a reserve blackout period.
    ///         The blackout period starts at RoundDuration - AuctionClosingSeconds - ReserveSubmissionSeconds,
    ///         and ends when the auction round is resolved, or the round ends.
    uint64 reserveSubmissionSeconds;
}

library RoundTimingInfoLib {
    /// @dev Using signed offset involves a lot of casting when comparing the to the block timestamp
    ///      so we provide a helper method here
    function blockTimestampBeforeOffset(
        int64 offsetTimestamp
    ) private view returns (bool) {
        return int64(uint64(block.timestamp)) < offsetTimestamp;
    }

    /// @dev Using signed offset involves a lot of casting when comparing the to the block timestamp
    ///      so we provide a helper method here
    ///      Notice! this helper method should not be used before checking that the offset is less than the timestamp
    function unsignedSinceTimestamp(
        int64 offsetTimestamp
    ) private view returns (uint64) {
        return uint64(int64(uint64(block.timestamp)) - offsetTimestamp);
    }

    /// @notice The current round, given the current timestamp, the offset and the round duration
    function currentRound(
        RoundTimingInfo memory info
    ) internal view returns (uint64) {
        return _currentRound(info.offsetTimestamp, info.roundDurationSeconds);
    }

    function currentRound(
        RoundTimingInfo storage info
    ) internal view returns (uint64) {
        return _currentRound(info.offsetTimestamp, info.roundDurationSeconds);
    }

    function currentRound(
        RoundTimingInfo calldata info
    ) internal view returns (uint64) {
        return _currentRound(info.offsetTimestamp, info.roundDurationSeconds);
    }

    /// @notice Has the current auction round closed
    function isAuctionRoundClosed(
        RoundTimingInfo memory info
    ) internal view returns (bool) {
        return _isAuctionRoundClosed(
            info.offsetTimestamp, info.roundDurationSeconds, info.auctionClosingSeconds
        );
    }

    function isAuctionRoundClosed(
        RoundTimingInfo storage info
    ) internal view returns (bool) {
        return _isAuctionRoundClosed(
            info.offsetTimestamp, info.roundDurationSeconds, info.auctionClosingSeconds
        );
    }

    function isAuctionRoundClosed(
        RoundTimingInfo calldata info
    ) internal view returns (bool) {
        return _isAuctionRoundClosed(
            info.offsetTimestamp, info.roundDurationSeconds, info.auctionClosingSeconds
        );
    }

    /// @notice The reserve cannot be set during the blackout period
    ///         This period runs from ReserveSubmissionSeconds before the auction closes and ends when the round resolves, or when the round ends.
    /// @param info Round timing info
    /// @param latestResolvedRound The last auction round number that was resolved
    function isReserveBlackout(
        RoundTimingInfo memory info,
        uint64 latestResolvedRound
    ) internal view returns (bool) {
        return _isReserveBlackout(
            info.offsetTimestamp,
            info.roundDurationSeconds,
            info.auctionClosingSeconds,
            info.reserveSubmissionSeconds,
            latestResolvedRound
        );
    }

    function isReserveBlackout(
        RoundTimingInfo storage info,
        uint64 latestResolvedRound
    ) internal view returns (bool) {
        return _isReserveBlackout(
            info.offsetTimestamp,
            info.roundDurationSeconds,
            info.auctionClosingSeconds,
            info.reserveSubmissionSeconds,
            latestResolvedRound
        );
    }

    function isReserveBlackout(
        RoundTimingInfo calldata info,
        uint64 latestResolvedRound
    ) internal view returns (bool) {
        return _isReserveBlackout(
            info.offsetTimestamp,
            info.roundDurationSeconds,
            info.auctionClosingSeconds,
            info.reserveSubmissionSeconds,
            latestResolvedRound
        );
    }

    /// @notice Gets the start and end timestamps (seconds) of a specified round
    ///         Since it is possible to set a negative offset, the start and end time may also be negative
    ///         In this case requesting roundTimestamps will revert.
    /// @param info Round timing info
    /// @param round The specified round
    /// @return The timestamp at which the round starts
    /// @return The timestamp at which the round ends
    function roundTimestamps(
        RoundTimingInfo memory info,
        uint64 round
    ) internal pure returns (uint64, uint64) {
        return _roundTimestamps(info.offsetTimestamp, info.roundDurationSeconds, round);
    }

    function roundTimestamps(
        RoundTimingInfo storage info,
        uint64 round
    ) internal view returns (uint64, uint64) {
        return _roundTimestamps(info.offsetTimestamp, info.roundDurationSeconds, round);
    }

    function roundTimestamps(
        RoundTimingInfo calldata info,
        uint64 round
    ) internal pure returns (uint64, uint64) {
        return _roundTimestamps(info.offsetTimestamp, info.roundDurationSeconds, round);
    }

    function _currentRound(
        int64 offsetTimestamp,
        uint64 roundDurationSeconds
    ) private view returns (uint64) {
        if (blockTimestampBeforeOffset(offsetTimestamp)) {
            return 0;
        }

        return unsignedSinceTimestamp(offsetTimestamp) / roundDurationSeconds;
    }

    function _isAuctionRoundClosed(
        int64 offsetTimestamp,
        uint64 roundDurationSeconds,
        uint64 auctionClosingSeconds
    ) private view returns (bool) {
        if (blockTimestampBeforeOffset(offsetTimestamp)) {
            return false;
        }

        uint64 timeInRound = unsignedSinceTimestamp(offsetTimestamp) % roundDurationSeconds;
        // round closes at AuctionClosedSeconds before the end of the round
        return timeInRound >= roundDurationSeconds - auctionClosingSeconds;
    }

    function _isReserveBlackout(
        int64 offsetTimestamp,
        uint64 roundDurationSeconds,
        uint64 auctionClosingSeconds,
        uint64 reserveSubmissionSeconds,
        uint64 latestResolvedRound
    ) private view returns (bool) {
        if (blockTimestampBeforeOffset(offsetTimestamp)) {
            // no rounds have started, can't be in blackout
            return false;
        }

        // if we're in round r, we are selling the rights for r+1
        // if the latest round is r+1 that means we've already resolved the auction in r
        // so we are no longer in the blackout period
        uint64 timeSinceOffset = unsignedSinceTimestamp(offsetTimestamp);
        uint64 curRound = timeSinceOffset / roundDurationSeconds;
        if (latestResolvedRound >= curRound + 1) {
            return false;
        }

        // the round in question hasnt been resolved
        // therefore if we're within ReserveSubmissionSeconds of the auction close then we're in blackout
        // otherwise we're not
        // reuse the quotient to derive the remainder without a second division
        uint64 timeInRound = timeSinceOffset - curRound * roundDurationSeconds;
        return timeInRound >= (roundDurationSeconds - auctionClosingSeconds - reserveSubmissionSeconds);
    }

    function _roundTimestamps(
        int64 offsetTimestamp,
        uint64 roundDurationSeconds,
        uint64 round
    ) private pure returns (uint64, uint64) {
        // Compute the offset in uint256 to catch overflow before truncating to int64.
        uint256 roundOffset = uint256(roundDurationSeconds) * uint256(round);
        require(roundOffset <= uint256(type(int64).max), "Round offset overflows int64");
        int64 intRoundStart = offsetTimestamp + int64(uint64(roundOffset));
        if (intRoundStart < 0) {
            revert NegativeRoundStart(intRoundStart);
        }
        uint64 roundStart = uint64(intRoundStart);
        uint64 roundEnd = roundStart + roundDurationSeconds - 1;
        return (roundStart, roundEnd);
    }
}
