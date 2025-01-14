// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import { ClonesWithImmutableArgs } from "@cwia/ClonesWithImmutableArgs.sol";
import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ISemver } from "src/universal/ISemver.sol";

import { IDisputeGame } from "src/dispute/interfaces/IDisputeGame.sol";
import { IDisputeGameFactory } from "src/dispute/interfaces/IDisputeGameFactory.sol";

import { LibGameId } from "src/dispute/lib/LibGameId.sol";

import "src/libraries/DisputeTypes.sol";
import "src/libraries/DisputeErrors.sol";

/// @title DisputeGameFactory
/// @notice A factory contract for creating `IDisputeGame` contracts. All created dispute games
///         are stored in both a mapping and an append only array. The timestamp of the creation
///         time of the dispute game is packed tightly into the storage slot with the address of
///         the dispute game. This is to make offchain discoverability of playable dispute games
///         easier.
contract DisputeGameFactory is OwnableUpgradeable, IDisputeGameFactory, ISemver {
    /// @dev Allows for the creation of clone proxies with immutable arguments.
    using ClonesWithImmutableArgs for address;

    /// @inheritdoc IDisputeGameFactory
    mapping(GameType => IDisputeGame) public gameImpls;

    /// @notice Mapping of a hash of `gameType || rootClaim || extraData` to
    ///         the deployed `IDisputeGame` clone.
    /// @dev Note: `||` denotes concatenation.
    mapping(Hash => GameId) internal _disputeGames;

    /// @notice an append-only array of disputeGames that have been created.
    /// @dev this accessor is used by offchain game solvers to efficiently
    ///      track dispute games
    GameId[] internal _disputeGameList;

    /// @notice Semantic version.
    /// @custom:semver 0.0.7
    string public constant version = "0.0.7";

    /// @notice constructs a new DisputeGameFactory contract.
    constructor() OwnableUpgradeable() {
        initialize(address(0));
    }

    /// @notice Initializes the contract.
    /// @param _owner The owner of the contract.
    function initialize(address _owner) public initializer {
        __Ownable_init();
        _transferOwnership(_owner);
    }

    /// @inheritdoc IDisputeGameFactory
    function gameCount() external view returns (uint256 gameCount_) {
        gameCount_ = _disputeGameList.length;
    }

    /// @inheritdoc IDisputeGameFactory
    function games(
        GameType _gameType,
        Claim _rootClaim,
        bytes calldata _extraData
    )
        external
        view
        returns (IDisputeGame proxy_, Timestamp timestamp_)
    {
        Hash uuid = getGameUUID(_gameType, _rootClaim, _extraData);
        (, timestamp_, proxy_) = _disputeGames[uuid].unpack();
    }

    /// @inheritdoc IDisputeGameFactory
    function gameAtIndex(uint256 _index)
        external
        view
        returns (GameType gameType_, Timestamp timestamp_, IDisputeGame proxy_)
    {
        (gameType_, timestamp_, proxy_) = _disputeGameList[_index].unpack();
    }

    /// @inheritdoc IDisputeGameFactory
    function create(
        GameType _gameType,
        Claim _rootClaim,
        bytes calldata _extraData
    )
        external
        returns (IDisputeGame proxy_)
    {
        // Grab the implementation contract for the given `GameType`.
        IDisputeGame impl = gameImpls[_gameType];

        // If there is no implementation to clone for the given `GameType`, revert.
        if (address(impl) == address(0)) revert NoImplementation(_gameType);

        // Clone the implementation contract and initialize it with the given parameters.
        proxy_ = IDisputeGame(address(impl).clone(abi.encodePacked(_rootClaim, _extraData)));
        proxy_.initialize();

        // Compute the unique identifier for the dispute game.
        Hash uuid = getGameUUID(_gameType, _rootClaim, _extraData);

        // If a dispute game with the same UUID already exists, revert.
        if (GameId.unwrap(_disputeGames[uuid]) != bytes32(0)) revert GameAlreadyExists(uuid);

        GameId id = LibGameId.pack(_gameType, Timestamp.wrap(uint64(block.timestamp)), proxy_);

        // Store the dispute game id in the mapping & emit the `DisputeGameCreated` event.
        _disputeGames[uuid] = id;
        _disputeGameList.push(id);
        emit DisputeGameCreated(address(proxy_), _gameType, _rootClaim);
    }

    /// @inheritdoc IDisputeGameFactory
    function getGameUUID(
        GameType _gameType,
        Claim _rootClaim,
        bytes calldata _extraData
    )
        public
        pure
        returns (Hash uuid_)
    {
        uuid_ = Hash.wrap(keccak256(abi.encode(_gameType, _rootClaim, _extraData)));
    }

    /// @inheritdoc IDisputeGameFactory
    function setImplementation(GameType _gameType, IDisputeGame _impl) external onlyOwner {
        gameImpls[_gameType] = _impl;
        emit ImplementationSet(address(_impl), _gameType);
    }
}
