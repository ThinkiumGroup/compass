// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	ZeroAddress = common.HexToAddress("0x0000000000000000000000000000000000000000")
	// swapIn( uint256 id, address token, address to, uint amount, uint fromChainID, address sourceRouter, bytes memory data)
	SwapIn = "swapIn(uint256,address,address,uint,uint,address,bytes)"
)

//var StoreFunctionSig = CreateFunctionSignature("store(bytes32)")

// CreateFunctionSignature hashes the function signature and returns the first 4 bytes
func CreateFunctionSignature(sig string) [4]byte {
	var res [4]byte
	hash := crypto.Keccak256Hash([]byte(sig))
	copy(res[:], hash[:])
	return res
}
