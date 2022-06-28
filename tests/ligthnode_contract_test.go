package tests

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"testing"
	"time"

	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	maptypes "github.com/mapprotocol/atlas/core/types"
	"github.com/neoiss/ethclient"

	"github.com/mapprotocol/compass/mapprotocol"
	utils "github.com/mapprotocol/compass/shared/ethereum"
)

var ContractAddr = common.HexToAddress("0xA7D3A66013DE32f0a44C92E337Af22C4344a2d62")

func dialConn() *ethclient.Client {
	conn, err := ethclient.Dial("https://ropsten.infura.io/v3/8cce6b470ad44fb5a3621aa34243647f")
	if err != nil {
		log.Fatalf("Failed to connect to the atlas: %v", err)
	}
	return conn
}

func dialMapConn() *ethclient.Client {
	conn, err := ethclient.Dial("http://18.142.54.137:7445")
	if err != nil {
		log.Fatalf("Failed to connect to the atlas: %v", err)
	}
	return conn
}

func TestLoadPrivate(t *testing.T) {
	path := "/Users/t/data/atlas-1/keystore/UTC--2022-06-07T04-22-55.836701000Z--f9803e9021e56e68662351fe43773934c4a276b8"
	password := ""
	addr, private := LoadPrivate(path, password)
	fmt.Println("============================== addr: ", addr)
	fmt.Printf("============================== private key: %x\n", crypto.FromECDSA(private))
}
func TestUpdateHeader(t *testing.T) {
	cli := dialConn()
	for i := 1; i < 21; i++ {
		number := int64(i * 1000)
		fmt.Println("============================== number: ", number)
		header, err := cli.MAPHeaderByNumber(context.Background(), big.NewInt(number))
		if err != nil {
			t.Fatalf(err.Error())
		}

		h := mapprotocol.ConvertHeader(header)
		aggPK, err := mapprotocol.GetAggPK(cli, header.Number, header.Extra)
		if err != nil {
			t.Fatalf(err.Error())
		}

		//printHeader(header)
		//printAggPK(aggPK)
		//_ = h

		input, err := mapprotocol.PackLightNodeInput(mapprotocol.MethodUpdateBlockHeader, h, aggPK)
		if err != nil {
			t.Fatalf(err.Error())
		}

		path := "/Users/xm/Desktop/WL/code/atlas/node-1/keystore/UTC--2022-06-15T07-51-25.301943000Z--e0dc8d7f134d0a79019bef9c2fd4b2013a64fcd6"
		password := "1234"
		from, private := LoadPrivate(path, password)
		if err := SendContractTransaction(cli, from, ContractAddr, nil, private, input); err != nil {
			t.Fatalf(err.Error())
		}
	}
}

func TestVerifyProofData(t *testing.T) {
	var (
		number = big.NewInt(106020)
		//number       = big.NewInt(4108)
		//number       = big.NewInt(55342)
		txIndex uint = 0
	)
	cli := dialMapConn()

	header, err := cli.MAPHeaderByNumber(context.Background(), number)
	if err != nil {
		t.Fatalf(err.Error())
	}

	txsHash, err := getTransactionsHashByBlockNumber(cli, number)
	if err != nil {
		t.Fatalf(err.Error())
	}
	receipts, err := getReceiptsByTxsHash(cli, txsHash)
	if err != nil {
		t.Fatalf(err.Error())
	}
	receipt, err := mapprotocol.GetTxReceipt(receipts[txIndex])

	proof, err := getProof(receipts, txIndex)
	if err != nil {
		t.Fatalf(err.Error())
	}

	aggPK, err := mapprotocol.GetAggPK(cli, header.Number, header.Extra)
	if err != nil {
		t.Fatalf(err.Error())
	}

	var key []byte
	key = rlp.AppendUint64(key[:0], uint64(txIndex))

	//fmt.Println("============================== number: ", number)
	//printHeader(header)
	//printAggPK(aggPK)
	//printReceipt(receipt)
	//fmt.Println("============================== KeyIndex: ", "0x"+common.Bytes2Hex(key))
	//printProof(proof)

	rp := mapprotocol.ReceiptProof{
		Header:   mapprotocol.ConvertHeader(header),
		AggPk:    aggPK,
		Receipt:  receipt,
		KeyIndex: key,
		Proof:    proof,
	}

	input, err := mapprotocol.PackLightNodeInput(mapprotocol.MethodVerifyProofData, rp)
	if err != nil {
		t.Fatalf(err.Error())
	}
	path := "/Users/xm/Desktop/WL/code/atlas/node-1/keystore/UTC--2022-06-15T07-51-25.301943000Z--e0dc8d7f134d0a79019bef9c2fd4b2013a64fcd6"
	password := "1234"
	from, _ := LoadPrivate(path, password)
	output, err := dialConn().CallContract(context.Background(), eth.CallMsg{From: from, To: &ContractAddr, Data: input}, nil)
	if err != nil {
		t.Fatalf(err.Error())
	}

	resp, err := mapprotocol.ABILightNode.Methods[mapprotocol.MethodVerifyProofData].Outputs.Unpack(output)
	if err != nil {
		t.Fatalf(err.Error())
	}

	ret := struct {
		Success bool
		Message string
	}{}
	if err := mapprotocol.ABILightNode.Methods[mapprotocol.MethodVerifyProofData].Outputs.Copy(&ret, resp); err != nil {
		t.Fatalf(err.Error())
	}

	fmt.Printf("============================== success: %v, message: %v\n", ret.Success, ret.Message)
}

func TestGetLog(t *testing.T) {
	//number       = big.NewInt(4108)
	//number       = big.NewInt(55342)
	query := buildQuery(common.HexToAddress("0xf03aDB732FBa8Fca38C00253B1A1aa72CCA026E6"),
		utils.SwapOut, big.NewInt(106020), big.NewInt(106020))

	// querying for logs
	logs, err := dialConn().FilterLogs(context.Background(), query)
	if err != nil {
		t.Fatalf("unable to Filter Logs: %s", err)
	}
	t.Logf("log len is %v", len(logs))
}

// buildQuery constructs a query for the bridgeContract by hashing sig to get the event topic
func buildQuery(contract common.Address, sig utils.EventSig, startBlock *big.Int, endBlock *big.Int) eth.FilterQuery {
	query := eth.FilterQuery{
		FromBlock: startBlock,
		ToBlock:   endBlock,
		Addresses: []common.Address{contract},
		Topics: [][]common.Hash{
			{sig.GetTopic()},
		},
	}
	return query
}

func SendContractTransaction(client *ethclient.Client, from, toAddress common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, input []byte) error {
	// Ensure a valid value field and resolve the account nonce
	nonce, err := client.PendingNonceAt(context.Background(), from)
	if err != nil {
		return err
	}
	//fmt.Printf("============================== from: %s, nonce: %d\n", from, nonce)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	gasLimit := uint64(2100000) // in units
	// If the contract surely has code (or code is not needed), estimate the transaction
	msg := eth.CallMsg{From: from, To: &toAddress, GasPrice: gasPrice, Value: value, Data: input}
	gasLimit, err = client.EstimateGas(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("contract exec failed, %s", err.Error())
	}
	if gasLimit < 1 {
		gasLimit = 866328
	}

	// Create the transaction, sign it and schedule it for execution
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, input)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return err
	}
	//fmt.Println("TX data nonce ", nonce, " transfer value ", value, " gasLimit ", gasLimit, " gasPrice ", gasPrice, " chainID ", chainID)
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}
	txHash := signedTx.Hash()
	count := 0
	for {
		time.Sleep(time.Millisecond * 1000)
		_, isPending, err := client.TransactionByHash(context.Background(), txHash)
		if err != nil {
			return err
		}
		count++
		if !isPending {
			break
		} else {
			log.Println("======================== pending...")
		}
	}
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return err
	}
	if receipt.Status == types.ReceiptStatusSuccessful {
		logs, _ := json.Marshal(receipt.Logs)
		log.Printf("Transaction Success, number: %v, hash: %v， logs: %v\n", receipt.BlockNumber.Uint64(), receipt.BlockHash, string(logs))
	} else if receipt.Status == types.ReceiptStatusFailed {
		log.Println("Transaction Failed. ", "block number: ", receipt.BlockNumber.Uint64())
		return errors.New("transaction failed")
	}
	return nil
}

func LoadPrivate(path, password string) (common.Address, *ecdsa.PrivateKey) {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	key, err := keystore.DecryptKey(bs, password)
	if err != nil || key == nil {
		panic(fmt.Errorf("error decrypting key: %v", err))
	}
	priKey := key.PrivateKey
	addr := crypto.PubkeyToAddress(priKey.PublicKey)

	if priKey == nil {
		panic("load privateKey failed")
	}
	return addr, priKey
}

func printHeader(header *maptypes.Header) {
	type blockHeader struct {
		ParentHash  string
		Coinbase    string
		Root        string
		TxHash      string
		ReceiptHash string
		Bloom       string
		Number      *big.Int
		GasLimit    *big.Int
		GasUsed     *big.Int
		Time        *big.Int
		ExtraData   string
		MixDigest   string
		Nonce       string
		BaseFee     *big.Int
	}
	h := blockHeader{
		ParentHash:  "0x" + common.Bytes2Hex(header.ParentHash[:]),
		Coinbase:    header.Coinbase.String(),
		Root:        "0x" + common.Bytes2Hex(header.Root[:]),
		TxHash:      "0x" + common.Bytes2Hex(header.TxHash[:]),
		ReceiptHash: "0x" + common.Bytes2Hex(header.ReceiptHash[:]),
		Bloom:       "0x" + common.Bytes2Hex(header.Bloom[:]),
		Number:      header.Number,
		GasLimit:    new(big.Int).SetUint64(header.GasLimit),
		GasUsed:     new(big.Int).SetUint64(header.GasUsed),
		Time:        new(big.Int).SetUint64(header.Time),
		ExtraData:   "0x" + common.Bytes2Hex(header.Extra),
		MixDigest:   "0x" + common.Bytes2Hex(header.MixDigest[:]),
		Nonce:       "0x" + common.Bytes2Hex(header.Nonce[:]),
		BaseFee:     header.BaseFee,
	}
	fmt.Printf("============================== header: %+v\n", h)
}

func printAggPK(aggPk *mapprotocol.G2) {
	type G2Str struct {
		xr string
		xi string
		yr string
		yi string
	}
	g2 := G2Str{
		xr: "0x" + common.Bytes2Hex(aggPk.Xr.Bytes()),
		xi: "0x" + common.Bytes2Hex(aggPk.Xi.Bytes()),
		yr: "0x" + common.Bytes2Hex(aggPk.Yr.Bytes()),
		yi: "0x" + common.Bytes2Hex(aggPk.Yi.Bytes()),
	}
	fmt.Printf("============================== aggPk: %+v\n", g2)
}

func printReceipt(r *mapprotocol.TxReceipt) {
	type txLog struct {
		Addr   common.Address
		Topics []string
		Data   string
	}

	type receipt struct {
		ReceiptType       *big.Int
		PostStateOrStatus string
		CumulativeGasUsed *big.Int
		Bloom             string
		Logs              []txLog
	}

	logs := make([]txLog, 0, len(r.Logs))
	for _, lg := range r.Logs {
		topics := make([]string, 0, len(lg.Topics))
		for _, tp := range lg.Topics {
			topics = append(topics, "0x"+common.Bytes2Hex(tp))
		}
		logs = append(logs, txLog{
			Addr:   lg.Addr,
			Topics: topics,
			Data:   "0x" + common.Bytes2Hex(lg.Data),
		})
	}

	rr := receipt{
		ReceiptType:       r.ReceiptType,
		PostStateOrStatus: "0x" + common.Bytes2Hex(r.PostStateOrStatus),
		CumulativeGasUsed: r.CumulativeGasUsed,
		Bloom:             "0x" + common.Bytes2Hex(r.Bloom),
		Logs:              logs,
	}
	fmt.Printf("============================== Receipt: %+v\n", rr)
}

func printProof(proof [][]byte) {
	p := make([]string, 0, len(proof))
	for _, v := range proof {
		p = append(p, "0x"+common.Bytes2Hex(v))
	}
	fmt.Println("============================== proof: ", p)
}

func getProof(receipts []*types.Receipt, txIndex uint) ([][]byte, error) {
	tr, err := trie.New(common.Hash{}, trie.NewDatabase(memorydb.New()))
	if err != nil {
		return nil, err
	}

	tr = utils.DeriveTire(receipts, tr)
	ns := light.NewNodeSet()
	key, err := rlp.EncodeToBytes(txIndex)
	if err != nil {
		return nil, err
	}
	if err = tr.Prove(key, 0, ns); err != nil {
		return nil, err
	}

	proof := make([][]byte, 0, len(ns.NodeList()))
	for _, v := range ns.NodeList() {
		proof = append(proof, v)
	}
	return proof, nil
}

func getTransactionsHashByBlockNumber(conn *ethclient.Client, number *big.Int) ([]common.Hash, error) {
	block, err := conn.MAPBlockByNumber(context.Background(), number)
	if err != nil {
		return nil, err
	}

	txs := make([]common.Hash, 0, len(block.Transactions()))
	for _, tx := range block.Transactions() {
		txs = append(txs, tx.Hash())
	}
	return txs, nil
}

func getReceiptsByTxsHash(conn *ethclient.Client, txsHash []common.Hash) ([]*types.Receipt, error) {
	rs := make([]*types.Receipt, 0, len(txsHash))
	for _, h := range txsHash {
		r, err := conn.TransactionReceipt(context.Background(), h)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
	return rs, nil
}