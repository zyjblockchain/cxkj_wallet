package logics

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/utils"
	"github.com/zyjblockchain/tt_tac/utils/ding_robot"
	transaction "github.com/zyjblockchain/tt_tac/utils/tx_utils"
	"math/big"
	"time"
)

type RespBalance struct {
	TtBalance  string `json:"tt_balance"`
	EthBalance string `json:"eth_balance"`
	Decimal    int    `json:"decimal"`
}
type GetBalance struct {
	Address string `json:"address" binding:"required"`
}

// GetBalance 获取主网余额
func (g *GetBalance) GetBalance() (*RespBalance, error) {
	ethClient := transaction.NewChainClient(conf.EthChainNet, big.NewInt(conf.EthChainID))
	defer ethClient.Close()

	EthBalance, err := ethClient.Client.BalanceAt(context.Background(), common.HexToAddress(g.Address), nil)
	if err != nil {
		log.Errorf("获取eth链上的eth币 balance error: %v, address: %s", err, g.Address)
		return nil, err
	}
	return &RespBalance{
		TtBalance:  "0.00",
		EthBalance: utils.UnitConversion(EthBalance.String(), 18, 6),
		Decimal:    18, // 这两个币的小数位数都是18位
	}, nil
}

type RespTokenBalance struct {
	TtPalaBalance  string `json:"tt_pala_balance"`
	EthPalaBalance string `json:"eth_pala_balance"`
	EthUsdtBalance string `json:"eth_usdt_balance"`
	UsdtDecimal    int    `json:"usdt_decimal"` // 6位小数
	PalaDecimal    int    `json:"pala_decimal"` // 18位
}
type TokenBalance struct {
	Address string `json:"address" binding:"required"`
}

// GetTokenBalance
func (t *TokenBalance) GetTokenBalance() (*RespTokenBalance, error) {
	ethClient := transaction.NewChainClient(conf.EthChainNet, big.NewInt(conf.EthChainID))
	defer ethClient.Close()

	EthPalaBalance, err := ethClient.GetTokenBalance(common.HexToAddress(t.Address), common.HexToAddress(conf.EthPalaTokenAddress))
	if err != nil {
		log.Errorf("获取eth pala balance err:%v. address: %s", err, t.Address)
		return nil, err
	}

	EthUsdtBalance, err := ethClient.GetTokenBalance(common.HexToAddress(t.Address), common.HexToAddress(conf.EthUSDTTokenAddress))
	if err != nil {
		log.Errorf("获取eth usdt balance err:%v. address: %s", err, t.Address)
		return nil, err
	}

	return &RespTokenBalance{
		TtPalaBalance:  "0.00",
		EthPalaBalance: utils.UnitConversion(EthPalaBalance.String(), 18, 6),
		EthUsdtBalance: utils.UnitConversion(EthUsdtBalance.String(), 6, 6),
		UsdtDecimal:    6,
		PalaDecimal:    18,
	}, nil
}

type GetGasFee struct {
	ChainTag int `json:"chain_tag" binding:"required"`
}

type Fee struct {
	GasFee string `json:"gas_fee"`
}

func (g *GetGasFee) GetGasFee() (*Fee, error) {
	var chainUrl string
	var chainId *big.Int

	chainUrl = conf.EthChainNet
	chainId = big.NewInt(conf.EthChainID)

	client := transaction.NewChainClient(chainUrl, chainId)
	defer client.Close()
	suggestPrice, err := client.SuggestGasPrice()
	if err != nil {
		log.Errorf("get suggest gas price err: %v", err)
		return nil, err
	}
	gasPrice := new(big.Int).Mul(suggestPrice, big.NewInt(2)) // 两倍于gasPrice
	log.Infof("gasPrice: %s", gasPrice.String())
	gasLimit := uint64(65000)
	gasFee := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit))).String()
	return &Fee{GasFee: utils.UnitConversion(gasFee, 18, 6)}, nil
}

// CheckMiddleAddressBalance 定时检查中间地址的各种资产的balance是否足够
func CheckMiddleAddressBalance() {
	dingRobot := ding_robot.NewRobot(conf.BalanceWebHook)
	log.Infof("中转地址balance钉钉告警webHook: %s", conf.BalanceWebHook)
	getBalanceTicker := time.NewTicker(30 * time.Second)
	ethClient := transaction.NewChainClient(conf.EthChainNet, big.NewInt(int64(conf.EthChainID)))
	defer func() {
		ethClient.Close()
	}()
	for {
		select {
		case <-getBalanceTicker.C:
			// 1. 查询闪兑中间地址的eth余额
			getFlashMiddleEthBalance, err := ethClient.Client.BalanceAt(context.Background(), common.HexToAddress(conf.EthFlashChangeMiddleAddress), nil)
			if err != nil {
				log.Errorf("查询闪兑中间地址的eth余额 error: %v", err)
			} else {
				// 最小余额限度0.5 eth
				limitBalance, _ := new(big.Int).SetString("500000000000000000", 10)
				if getFlashMiddleEthBalance.Cmp(limitBalance) < 0 {
					// 通知需要充eth了
					content := fmt.Sprintf("3.闪兑中转地址eth余额即将消耗完;\naddress: %s,\nbalance: %s eth", conf.EthFlashChangeMiddleAddress, utils.UnitConversion(getFlashMiddleEthBalance.String(), 18, 6))
					_ = dingRobot.SendText(content, nil, true)
				}
			}

			// 2. 查询闪兑中间地址的pala余额
			getFlashMiddleEthPalaBalance, err := ethClient.GetTokenBalance(common.HexToAddress(conf.EthFlashChangeMiddleAddress), common.HexToAddress(conf.EthPalaTokenAddress))
			if err != nil {
				log.Errorf("查询闪兑中间地址的以太坊上的pala余额 error: %v", err)
			} else {
				// 最小余额限度 1000 pala
				limitBalance, _ := new(big.Int).SetString("100000000000", 10)
				if getFlashMiddleEthPalaBalance.Cmp(limitBalance) < 0 {
					content := fmt.Sprintf("4.闪兑中转地址以太坊上的pala余额即将消耗完;\naddress: %s,\neth_pala_balance: %s eth", conf.EthFlashChangeMiddleAddress, utils.UnitConversion(getFlashMiddleEthPalaBalance.String(), 18, 6))
					_ = dingRobot.SendText(content, nil, true)
				}
			}
		}
	}
}
