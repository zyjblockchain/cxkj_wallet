package conf

const (

	// // 测试网
	// EthChainNet = "https://rinkeby.infura.io/v3/36b98a13557c4b8583d57934ede2f74d"
	// EthChainID  = 4

	// 备用节点
	// EthChainNet = "https://mainnet.infura.io/v3/36b98a13557c4b8583d57934ede2f74d" // 18382255942

	// 主网
	EthChainNet = "https://mainnet.infura.io/v3/7bbf73a8855d4c0491f93e6dc498360d" // 1263344073
	EthChainID  = 1                                                               // 以太坊的主网chainId == 1

	EthChainTag = 17
)

// 正式环境的token
const (
	EthPalaTokenAddress = "0x13056817f997bc3f15e1bc68207efe8d2d197308" // 以太坊主网上的cxkj合约地址
	EthUSDTTokenAddress = "0xdAC17F958D2ee523a2206206994597C13D831ec7" // 以太坊主网上的USDT合约地址
)

// // 测试环境的token
// const (
// 	EthPalaTokenAddress = "0x03332638A6b4F5442E85d6e6aDF929Cd678914f1" // 以太坊上的pala erc20 地址，目前是测试环境 以太坊rinkeby 上的test3
// 	EthUSDTTokenAddress = "0xD1Df5b185198F3c6Da74e93B36b7E29523c265F0" // 以太坊上的usdt erc20 地址, 目前是以太坊测试网的测试token
// )

// 跨链转账扣除pala手续费数量
var (
	FlashPalaToUsdtPriceChange = float64(1.01) // 在闪兑中pala的价格需要增大来展示给用户，通过这种方式变相收取闪兑的手续费。默认上浮1%，有接口可以随时修改
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
