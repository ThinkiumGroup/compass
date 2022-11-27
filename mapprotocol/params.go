// Copyright 2021 Compass Systems
// SPDX-License-Identifier: LGPL-3.0-only

package mapprotocol

import (
	"github.com/mapprotocol/compass/msg"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

const (
	MethodVerifyProofData       = "verifyProofData"
	MethodUpdateBlockHeader     = "updateBlockHeader"
	MethodVerifiableHeaderRange = "verifiableHeaderRange"
	MethodOfHeaderHeight        = "headerHeight"
	MethodOfTransferIn          = "transferIn"
	MethodOfDepositIn           = "depositIn"
	MethodOfRegister            = "register"
	MethodOfBindWorker          = "bind"
	MethodOfOrderList           = "orderList"
	MethodOfIsUsedEvent         = "is_used_event"
	MethodOfGetBytes            = "getBytes"
	MethodOfGetHeadersBytes     = "getHeadersBytes"
)

const (
	NearVerifyRange  = "get_verifiable_header_range"
	NearHeaderHeight = "get_header_height"
)

const (
	EpochOfMap         = 50000
	EpochOfBsc         = 200
	HeaderCountOfBsc   = 12
	HeaderCountOfMatic = 64
)

// common varible
var (
	Big0           = big.NewInt(0)
	Big1           = big.NewInt(1)
	RegisterAmount = int64(100) // for test purpose
)

var (
	ZeroAddress      = common.HexToAddress("0x0000000000000000000000000000000000000000")
	RelayerAddress   = common.HexToAddress("0x000068656164657273746F726541646472657373")
	HashOfTransferIn = common.HexToHash("0x44ff77018688dad4b245e8ab97358ed57ed92269952ece7ffd321366ce078622")
	NearOfTransferIn = "2ef1cdf83614a69568ed2c96a275dd7fb2e63a464aa3a0ffe79f55d538c8b3b5"
)

var (
	Mcs, _         = abi.JSON(strings.NewReader(McsAbi))
	Bsc, _         = abi.JSON(strings.NewReader(BscAbiJson))
	Near, _        = abi.JSON(strings.NewReader(NearAbiJson))
	LightManger, _ = abi.JSON(strings.NewReader(LightMangerAbi))
	Map2Other, _   = abi.JSON(strings.NewReader(Map2OtherAbi))
	ABIRelayer, _  = abi.JSON(strings.NewReader(RelayerABIJSON))
	Height, _      = abi.JSON(strings.NewReader(HeightAbiJson))
	Verify, _      = abi.JSON(strings.NewReader(VerifiableHeaderRangeAbiJson))
	Matic, _       = abi.JSON(strings.NewReader(MaticAbiJson))
)

type Role string

var (
	RoleOfMaintainer Role = "maintainer"
	RoleOfMessenger  Role = "messenger"
)

var (
	OnlineChaId = map[msg.ChainId]string{}
)
