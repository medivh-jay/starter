package payments

import (
	"net/http"
	"starter/pkg/app"
	"sync"
)

// CreateResultParams 创建订单成功之后返回的数据结构
type CreateResultParams struct {
	State         bool   `json:"state"`          // 订单创建状态
	PaymentID     string `json:"payment_id"`     // 创建成功的话，就是服务器的唯一订单号,需要保存
	QrCode        string `json:"qr_code"`        // 可能会有二维码，这里就是二维码的字符串信息
	PaymentLink   string `json:"payment_link"`   // 可能会有充值链接,这里就是充值链接，直接打开就调起来充值的那种
	PaymentParams string `json:"payment_params"` // 可能某些充值平台在手机上调起充值方式比较妖孽，需要一串参数，就是这个
	ErrMsg        string `json:"err_msg"`
}

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Payment)
)

// Payment 支付对象实现该接口方法
type Payment interface {
	/* 对渠道商创建一笔支付订单 */
	/* 参数 */
	/*   id: 发起充值时声称的订单ID */
	/*   title: 发起充值时声称的订单ID */
	/*   body: 发起充值时声称的订单ID */
	/* 返回值 */
	/*   state: 充值订单是否创建成功
	/*   paymentId: 充值订单id */
	/*   params: 充值参数,不同充值平台参数可能不一致 */
	Create(id, title, body string, total int) CreateResultParams

	/* 检查充值是否成功 */
	/* 参数 */
	/*   paymentId: 发起充值时声称的订单ID */
	/* 返回值 */
	/*   state: 充值是否成功
	/*   total: 充值成功的话，充值成功了多少钱，单位：分 */
	CheckRechargeState(paymentID string) (state bool, total int)

	/* 渠道异步通知 */
	/* 参数 */
	/*   data: 通知的数据参数 */
	/* 返回值 */
	/*   state: 充值是否成功 */
	/*   paymentId: 订单id */
	/*   total: 充值成功的话，充值成功了多少钱，单位：分 */
	Notify(rep http.ResponseWriter, req *http.Request) (state bool, paymentID string, total int)

	// 这个接口是用来处理充值平台的异步回调的
	AckNotify(rep http.ResponseWriter)

	/* 退款 */
	/*    退款只能对充值订单发起退款申请,资金将会按充值原路退回 */
	/* 参数 */
	/*   paymentId: 订单id */
	Refund(paymentID string) interface{}

	// APP 支付调用
	SdkPay(id, title, body string, total int) CreateResultParams
}

// Register 将自己实现的支付操作注册进来,方便统一调用
func Register(name string, driver Payment) {
	driversMu.Lock()
	defer driversMu.Unlock()

	if _, dup := drivers[name]; dup {
		panic("payment: Register called twice for driver " + name)
	}

	drivers[name] = driver
}

// Get 根据注册时传入的name获取对象
func Get(name string) Payment {
	if driver, ok := drivers[name]; ok {
		return driver
	}
	app.Logger().WithField("log_type", "pkg.payments.payments").Error("payment: driver does not exists")
	return nil
}
