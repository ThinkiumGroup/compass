package chain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	gconfig "github.com/mapprotocol/compass/config"
	"github.com/mapprotocol/compass/connections/ethereum/egs"
	"github.com/mapprotocol/compass/core"
	"github.com/mapprotocol/compass/msg"
	utils "github.com/mapprotocol/compass/shared/ethereum"
)

const (
	DefaultGasLimit           = 6721975
	DefaultGasPrice           = 20000000000
	DefaultBlockConfirmations = 20
	DefaultGasMultiplier      = 1
)

// Chain specific options
var (
	McsOpt                = "mcs"
	MaxGasPriceOpt        = "maxGasPrice"
	GasLimitOpt           = "gasLimit"
	GasMultiplier         = "gasMultiplier"
	LimitMultiplier       = "limitMultiplier"
	HttpOpt               = "http"
	StartBlockOpt         = "startBlock"
	BlockConfirmationsOpt = "blockConfirmations"
	EGSApiKey             = "egsApiKey"
	EGSSpeed              = "egsSpeed"
	SyncToMap             = "syncToMap"
	SyncIDList            = "syncIdList"
	LightNode             = "lightnode"
	Event                 = "event"
	WaterLine             = "waterLine"
	ChangeInterval        = "changeInterval"
	Eth2Url               = "eth2Url"
)

// Config encapsulates all necessary parameters in ethereum compatible forms
type Config struct {
	Name               string      // Human-readable chain name
	Id                 msg.ChainId // ChainID
	Endpoint           string      // url for rpc endpoint
	From               string      // address of key to use
	KeystorePath       string      // Location of keyfiles
	BlockstorePath     string
	FreshStart         bool // Disables loading from blockstore at start
	McsContract        common.Address
	GasLimit           *big.Int
	MaxGasPrice        *big.Int
	GasMultiplier      float64
	LimitMultiplier    float64
	Http               bool // Config for type of connection
	StartBlock         *big.Int
	BlockConfirmations *big.Int
	EgsApiKey          string // API key for ethgasstation to query gas prices
	EgsSpeed           string // The speed which a transaction should be processed: average, fast, fastest. Default: fast
	SyncToMap          bool   // Whether sync blockchain headers to Map
	MapChainID         msg.ChainId
	SyncChainIDList    []msg.ChainId  // chain ids which map sync to
	LightNode          common.Address // the lightnode to sync header
	SyncMap            map[msg.ChainId]*big.Int
	Events             []utils.EventSig
	SkipError          bool
	HooksUrl           string
	WaterLine          string
	ChangeInterval     string
	Eth2Endpoint       string
}

// ParseConfig uses a core.ChainConfig to construct a corresponding Config
func ParseConfig(chainCfg *core.ChainConfig) (*Config, error) {
	config := &Config{
		Name:               chainCfg.Name,
		Id:                 chainCfg.Id,
		Endpoint:           chainCfg.Endpoint,
		From:               chainCfg.From,
		KeystorePath:       chainCfg.KeystorePath,
		BlockstorePath:     chainCfg.BlockstorePath,
		FreshStart:         chainCfg.FreshStart,
		McsContract:        utils.ZeroAddress,
		GasLimit:           big.NewInt(DefaultGasLimit),
		MaxGasPrice:        big.NewInt(DefaultGasPrice),
		GasMultiplier:      DefaultGasMultiplier,
		LimitMultiplier:    DefaultGasMultiplier,
		Http:               false,
		StartBlock:         big.NewInt(0),
		BlockConfirmations: big.NewInt(0),
		EgsApiKey:          "",
		EgsSpeed:           "",
		Events:             make([]utils.EventSig, 0),
		SkipError:          chainCfg.SkipError,
		WaterLine:          "",
		ChangeInterval:     "",
		Eth2Endpoint:       "",
	}

	if contract, ok := chainCfg.Opts[McsOpt]; ok && contract != "" {
		config.McsContract = common.HexToAddress(contract)
		delete(chainCfg.Opts, McsOpt)
	} else {
		return nil, fmt.Errorf("must provide opts.mcs field for ethereum config")
	}

	if gasPrice, ok := chainCfg.Opts[MaxGasPriceOpt]; ok {
		price := big.NewInt(0)
		_, pass := price.SetString(gasPrice, 10)
		if pass {
			config.MaxGasPrice = price
			delete(chainCfg.Opts, MaxGasPriceOpt)
		} else {
			return nil, errors.New("unable to parse max gas price")
		}
	}

	if gasLimit, ok := chainCfg.Opts[GasLimitOpt]; ok {
		limit := big.NewInt(0)
		_, pass := limit.SetString(gasLimit, 10)
		if pass {
			config.GasLimit = limit
			delete(chainCfg.Opts, GasLimitOpt)
		} else {
			return nil, errors.New("unable to parse gas limit")
		}
	}

	if gasMultiplier, ok := chainCfg.Opts[GasMultiplier]; ok {
		float, err := strconv.ParseFloat(gasMultiplier, 64)
		if err == nil {
			config.GasMultiplier = float
			delete(chainCfg.Opts, GasMultiplier)
		} else {
			return nil, errors.New("unable to parse gasMultiplier to float")
		}
	}

	if limitMultiplier, ok := chainCfg.Opts[LimitMultiplier]; ok {
		float, err := strconv.ParseFloat(limitMultiplier, 64)
		if err == nil {
			config.LimitMultiplier = float
			delete(chainCfg.Opts, LimitMultiplier)
		} else {
			return nil, errors.New("unable to parse limitMultiplier to float")
		}
	}

	if HTTP, ok := chainCfg.Opts[HttpOpt]; ok && HTTP == "true" {
		config.Http = true
		delete(chainCfg.Opts, HttpOpt)
	} else if HTTP, ok := chainCfg.Opts[HttpOpt]; ok && HTTP == "false" {
		config.Http = false
		delete(chainCfg.Opts, HttpOpt)
	}

	if startBlock, ok := chainCfg.Opts[StartBlockOpt]; ok && startBlock != "" {
		block := big.NewInt(0)
		_, pass := block.SetString(startBlock, 10)
		if pass {
			config.StartBlock = block
			delete(chainCfg.Opts, StartBlockOpt)
		} else {
			return nil, fmt.Errorf("unable to parse %s", StartBlockOpt)
		}
	}

	if blockConfirmations, ok := chainCfg.Opts[BlockConfirmationsOpt]; ok && blockConfirmations != "" {
		val := big.NewInt(DefaultBlockConfirmations)
		_, pass := val.SetString(blockConfirmations, 10)
		if pass {
			config.BlockConfirmations = val
			delete(chainCfg.Opts, BlockConfirmationsOpt)
		} else {
			return nil, fmt.Errorf("unable to parse %s", BlockConfirmationsOpt)
		}
	} else {
		config.BlockConfirmations = big.NewInt(DefaultBlockConfirmations)
		delete(chainCfg.Opts, BlockConfirmationsOpt)
	}

	if gsnApiKey, ok := chainCfg.Opts[EGSApiKey]; ok && gsnApiKey != "" {
		config.EgsApiKey = gsnApiKey
		delete(chainCfg.Opts, EGSApiKey)
	}

	if speed, ok := chainCfg.Opts[EGSSpeed]; ok && speed == egs.Average || speed == egs.Fast || speed == egs.Fastest {
		config.EgsSpeed = speed
		delete(chainCfg.Opts, EGSSpeed)
	} else {
		// Default to "fast"
		config.EgsSpeed = egs.Fast
		delete(chainCfg.Opts, EGSSpeed)
	}

	if syncToMap, ok := chainCfg.Opts[SyncToMap]; ok && syncToMap == "true" {
		config.SyncToMap = true
		delete(chainCfg.Opts, SyncToMap)
	} else {
		delete(chainCfg.Opts, SyncToMap)
	}

	if mapChainID, ok := chainCfg.Opts[gconfig.MapChainID]; ok {
		// key exist anyway
		chainId, errr := strconv.Atoi(mapChainID)
		if errr != nil {
			return nil, errr
		}
		config.MapChainID = msg.ChainId(chainId)
		delete(chainCfg.Opts, gconfig.MapChainID)
	}

	if syncIDList, ok := chainCfg.Opts[SyncIDList]; ok && syncIDList != "[]" {
		err := json.Unmarshal([]byte(syncIDList), &config.SyncChainIDList)
		if err != nil {
			return nil, err
		}
		delete(chainCfg.Opts, SyncIDList)
	}

	if lightnode, ok := chainCfg.Opts[LightNode]; ok && lightnode != "" {
		config.LightNode = common.HexToAddress(lightnode)
	}

	if waterLine, ok := chainCfg.Opts[WaterLine]; ok && waterLine != "" {
		config.WaterLine = waterLine
	}

	if alarmSecond, ok := chainCfg.Opts[ChangeInterval]; ok && alarmSecond != "" {
		config.ChangeInterval = alarmSecond
	}

	if v, ok := chainCfg.Opts[Event]; ok && v != "" {
		vs := strings.Split(v, "|")
		for _, s := range vs {
			config.Events = append(config.Events, utils.EventSig(s))
		}
	}

	if eth2Url, ok := chainCfg.Opts[Eth2Url]; ok && eth2Url != "" {
		config.Eth2Endpoint = eth2Url
	}

	config.HooksUrl = os.Getenv("hooks")

	return config, nil
}
