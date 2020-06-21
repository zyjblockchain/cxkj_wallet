package conf

const (

	// // 测试网
	EthChainNet = "https://rinkeby.infura.io/v3/36b98a13557c4b8583d57934ede2f74d"
	EthChainID  = 4

	// 主网
	// EthChainNet = "https://mainnet.infura.io/v3/f1301efa5af1432c84063f231f08f920" // sandy@token.im
	// EthChainID  = 1                                                               // 以太坊的主网chainId == 1

	// 链的tag
	EthChainTag = 17
)

// // 正式环境的token
// const (
// 	EthPalaTokenAddress = "0x13056817f997bc3f15e1bc68207efe8d2d197308" // 以太坊主网上的cxkj合约地址
// 	EthUSDTTokenAddress = "0xdAC17F958D2ee523a2206206994597C13D831ec7" // 以太坊主网上的USDT合约地址
// )

// 测试环境的token
const (
	EthPalaTokenAddress = "0x11d0da63212d97060dae59cf33cb92da196308b7" // 以太坊上的pala erc20 地址，目前是测试环境 以太坊rinkeby 上的test3
	EthUSDTTokenAddress = "0xe3d152933bcc150ccd9c4accaefae499bc9ec80d" // 以太坊上的usdt erc20 地址, 目前是以太坊测试网的测试token
)

// 跨链转账扣除pala手续费数量
var (
	FlashPalaToUsdtPriceChange = float64(0.15) // 兑换比例，有接口可以随时修改
)

// 需要配置文件读取
var (
	Dsn = ""

	// eth usdt -> pala闪兑中转地址
	EthFlashChangeMiddleAddress = ""
	EthFlashChangeMiddlePrivate = ""
)
var BalanceWebHook = ""     // 中转地址余额不足的钉钉告警webHook
var AbnormalWebHook = ""    // 其他异常的钉钉告警webHook
var ReceiveUSDTAddress = "" // usdt归集接收地址
