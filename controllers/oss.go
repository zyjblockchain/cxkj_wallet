package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/serializer"
	"github.com/zyjblockchain/tt_tac/utils"
)

// EncryptoPrivate 对传入的私钥进行加密
type Priv struct {
	Private string `json:"private" binding:"required"`
}

func EncryptoPrivate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var p Priv
		err := c.ShouldBind(&p)
		if err != nil {
			log.Errorf("EncryptoPrivate should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}

		private := p.Private
		ePrivate, err := utils.EncryptPrivate(private)
		if err != nil {
			log.Errorf("EncryptPrivate  err: %v", err)
			serializer.ErrorResponse(c, utils.EncryptoPrivErrCode, utils.EncryptoPrivErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, ePrivate, "success")
		}
	}
}

// 查看闪兑中展示pala价格的上浮比例
func GetPalaPriceComeUpRate() gin.HandlerFunc {
	return func(c *gin.Context) {
		serializer.SuccessResponse(c, conf.FlashPalaToUsdtPriceChange, "success")
	}
}

// 上浮比例
type ratio struct {
	Rate string `json:"rate" binding:"required"`
}

// ModifyPalaPriceComeUpRate 修改闪兑中展示pala价格的上浮比例
func ModifyPalaPriceComeUpRate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var rt ratio
		err := c.ShouldBind(&rt)
		if err != nil {
			log.Errorf("ModifyPalaPriceComeUpRate should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}

		// 修改
		// 首先string转换成float64
		de, err := decimal.NewFromString(rt.Rate)
		if err != nil {
			log.Errorf("decimal.NewFromString(rt.Rate) err: %v", err)
			serializer.ErrorResponse(c, utils.ModifyPalaPriceComeUpRateErrCode, utils.ModifyPalaPriceComeUpRateErrMsg, err.Error())
			return
		}
		// 判断newVal是否小于1，如果小于1则不修改
		if de.Cmp(decimal.NewFromInt(1)) < 0 {
			serializer.SuccessResponse(c, nil, "不能输入小于1的上浮比例，系统默认是1.01，表示价格上浮1%")
			return
		}

		newVal, _ := de.Float64()
		// 重载
		serializer.SuccessResponse(c, nil, "success")
		conf.FlashPalaToUsdtPriceChange = newVal
		return
	}
}

// 获取闪兑的交易gas消耗总量
func GetFlashTotalGasFee() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 查询出闪兑订单表中的所有完成的订单记录的sendTxId
		var sendTxIds []uint
		if err := models.DB.Model(models.FlashChangeOrder{}).Where("state = ?", 1).Pluck("receive_tx_id", &sendTxIds).Error; err != nil {
			log.Errorf("从flashChangeOrder 表中拉取所有的完成的订单sendTxId失败： %v", err)
			serializer.SuccessResponse(c, nil, "")
			return
		}
		// 通过sendTxId查询出对应的交易的gasPrice
		var totalPrice = decimal.NewFromInt(0)
		for _, v := range sendTxIds {
			var tx = models.TxTransfer{}
			tx.ID = v
			models.DB.Select("gas_price").Take(&tx)
			// add
			d, _ := decimal.NewFromString(tx.GasPrice)
			totalPrice = d.Add(totalPrice)
		}
		// gas fee = gasPrice * gasLimit todo 这里默认gasLimit为60000，可能会有一点点误差，但可以忽略
		gasFee := totalPrice.Mul(decimal.NewFromInt(60000)).String()
		serializer.SuccessResponse(c, utils.UnitConversion(gasFee, 18, 6), "")
	}
}

type resp struct {
	PalaTotal string `json:"pala_total"`
	UsdtTotal string `json:"usdt_total"`
}

// GetFlashUsdtAndPalaTotalAmount
func GetFlashUsdtAndPalaTotalAmount() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 查询出闪兑订单表中的所有完成的订单记录的sendTxId
		var ords []models.FlashChangeOrder
		models.DB.Select("from_token_amount, to_token_amount").Where("state = ?", 1).Find(&ords)
		var fromTokenAmount = decimal.NewFromInt(0)
		var toTokenAmount = decimal.NewFromInt(0)
		for _, o := range ords {
			f, _ := decimal.NewFromString(o.FromTokenAmount) // 不重要的接口，忽略error
			t, _ := decimal.NewFromString(o.ToTokenAmount)
			fromTokenAmount = fromTokenAmount.Add(f)
			toTokenAmount = toTokenAmount.Add(t)
		}
		resp := resp{
			PalaTotal: utils.UnitConversion(fromTokenAmount.String(), 18, 6),
			UsdtTotal: utils.UnitConversion(toTokenAmount.String(), 6, 6),
		}
		serializer.SuccessResponse(c, resp, "success")
	}
}
