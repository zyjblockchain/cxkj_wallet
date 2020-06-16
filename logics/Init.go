package logics

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/gorm"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/models"
	transaction "github.com/zyjblockchain/tt_tac/utils/tx_utils"
	"math/big"
	"strings"
	"time"
)

// InitTacOrderState 服务重启之后闪兑的订单状态
func InitFlashOrderState(flashSvr *WatchFlashChange) {
	ethClient := transaction.NewChainClient(conf.EthChainNet, big.NewInt(conf.EthChainID))
	defer ethClient.Close()
	// 获取闪兑订单中状态为pending的记录
	flashOrds, err := new(models.FlashChangeOrder).GetFlashOrdersByState(0)
	if err != nil {
		log.Errorf("获取FlashChangeOrder表中状态为0的记录失败；error: %v", err)
		return
	}
	log.Infof("开始遍历查询出来的flash change order。订单数量：%d", len(flashOrds))
	for _, ord := range flashOrds {
		// 记录到缓存中
		FlashAddressMap[strings.ToLower(ord.OperateAddress)] = 1
		// 1. 查看是否有SendTxId
		if ord.SendTxId == 0 {
			log.Infof("1. 闪兑订单中没有sendTxId,则删除订单")
			// 直接删除订单
			_ = ord.Delete(ord.ID)
			continue
		}
		// 2. 存在则查看是否上链成功
		txTr, err := (&models.TxTransfer{Model: gorm.Model{ID: ord.SendTxId}}).Get()
		if err != nil {
			log.Errorf("2. 通过闪兑订单中的sendTxId查询txTransfer失败")
			// 直接删除订单
			_ = ord.Delete(ord.ID)
			continue
		}
		sendTxHash := txTr.TxHash
		log.Infof("2.2 闪兑sendTxHash: %s", sendTxHash)
		_, isPending, err := ethClient.Client.TransactionByHash(context.Background(), common.HexToHash(sendTxHash))
		if err != nil {
			log.Errorf("2.3 闪兑申请者send的usdt交易没有上链, err: %v", err)
			_ = ord.Delete(ord.ID)
			continue
		} else if isPending {
			// 2.1 此交易正在pending
			log.Infof("2.4 闪兑申请者发送的交易正在pending,开启交易监听; txHash: %s", sendTxHash)
			// 开启协程轮询监听交易情况
			go func() {
				var count = 5
				for {
					count--
					time.Sleep(20 * time.Second)
					_, isPending, err := ethClient.Client.TransactionByHash(context.Background(), common.HexToHash(sendTxHash))
					if err == nil && !isPending { // 查询到交易
						log.Infof("2.5 监听到闪兑申请者发送的usdt交易，开启发送pala给申请者; txHash: %s", sendTxHash)
						// 则执行闪兑中间地址转账部分
						if err := flashSvr.ProcessCollectFlashChangeTx(ord.OperateAddress, ord.FromTokenAmount); err != nil {
							log.Errorf("flashSvr.ProcessCollectFlashChangeTx(ord.OperateAddress, ord.FromTokenAmount) error: %v", err)
						}
						return
					}

					if count == 0 {
						// 设置订单为失败，并退出
						log.Errorf("监听申请者的闪兑发送的usdt交易失败，闪兑订单置位失败状态。txHash: %s", sendTxHash)
						_ = ord.Update(models.FlashChangeOrder{State: 2})
						return
					}
				}
			}()

			continue
		} else {
			log.Infof("3.0 闪兑sendTxHash链上查询到了；hash: %s", sendTxHash)
			// 3. 查询到了交易，则查看中间地址是否转了pala完成了闪兑的后半部分
			// 3.1 中间地址没有开始转 pala交易， 则执行后半部分事务
			if ord.ReceiveTxId == 0 {
				// 则执行闪兑中间地址转账部分
				log.Infof("3.1 闪兑中间地址发送pala给申请者账户；address: %s, usdtAmount: %s", ord.OperateAddress, ord.FromTokenAmount)
				if err := flashSvr.ProcessCollectFlashChangeTx(ord.OperateAddress, ord.FromTokenAmount); err != nil {
					log.Errorf("flashSvr.ProcessCollectFlashChangeTx(ord.OperateAddress, ord.FromTokenAmount) error: %v", err)
				}
				continue
			} else {
				// 3.2 存在则查看交易详情
				txTr, err := (&models.TxTransfer{Model: gorm.Model{ID: ord.ReceiveTxId}}).Get()
				if err != nil {
					log.Errorf("3.2 通过闪兑订单中的ReceiveTxId查询txTransfer失败")
					// 直接删除订单
					_ = ord.Delete(ord.ID)
					continue
				}
				receiveTxHash := txTr.TxHash
				_, _, err = ethClient.Client.TransactionByHash(context.Background(), common.HexToHash(receiveTxHash))
				if err == ethereum.NotFound {
					log.Infof("3.3 闪兑 receiveTxHash：%s 在链上查询不到，则开启中间地址发送pala给申请者账户address: %s, usdtAmount: %s", receiveTxHash, ord.OperateAddress, ord.FromTokenAmount)
					// 表示交易发送失败则中转地址需要重新发送
					if err := flashSvr.ProcessCollectFlashChangeTx(ord.OperateAddress, ord.FromTokenAmount); err != nil {
						log.Errorf("flashSvr.ProcessCollectFlashChangeTx(ord.OperateAddress, ord.FromTokenAmount) error: %v", err)
					}
					continue
				} else {
					log.Infof("3.4 闪兑订单满足完成条件，把订单置位成功状态；flashOrderId: %d", ord.ID)
					// 设置订单状态为成功状态 todo 只要中间地址成功发送交易上链则代表一定会转账成功
					_ = ord.Update(models.FlashChangeOrder{State: 1})
					continue
				}
			}
		}
	}
}
