package middlewares

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"starter/pkg/app"
	"starter/pkg/config"
	"starter/pkg/server"
)

type claims struct {
	Id        interface{} // 唯一id
	CheckData string      // 验证信息
	jwt.StandardClaims
}

//// 认证表信息,该表需要为mongo
//type Auth struct {
//	Entity  app.Table // 非指针对象
//	ParseId func(string) interface{}
//}

// 参与 jwt 数据表结构体需要实现这些接口
type AuthInterface interface {
	app.Table
	GetTopic() interface{}                         // 返回唯一信息
	FindByTopic(topic interface{}) AuthInterface   // 根据唯一信息标识获取数据信息, 比如根据用户id获取用户信息,需要注意传入的数据类型
	GetCheckData() string                          // 获取验证信息, jwt加密时, 改信息会一起进行加密, 解密时会解出来然后调用 Check 验证该信息的正确性, 如果是其他数据类型直接转string，比如是个结构体或者map, 直接转为json string
	Check(ctx *gin.Context, checkData string) bool // 验证信息
	ExpiredAt() int64                              // 返回过期时间,时间戳
}

var AuthEntity AuthInterface
var (
	AuthKey      = "users"
	JwtHeaderKey = "JWT"
)

// 验证用户有效性的中间件
func VerifyAuth(c *gin.Context) {
	token := c.GetHeader(JwtHeaderKey)
	if token != "" {
		claims, err := ParseToken(token)
		if err == nil {
			var entity = AuthEntity.FindByTopic(claims.Id)
			if entity.Check(c, claims.CheckData) {
				c.Set(AuthKey, entity) // 向下设置用户信息,控制器可直接获取
				c.Header(JwtHeaderKey, token)
				c.Next()
				return
			} else {
				app.NewResponse(app.AuthFail, nil, app.AuthFailMessage).End(c, http.StatusUnauthorized)
				c.Abort()
				return
			}
		}
	}
	app.NewResponse(app.AuthFail, nil, app.AuthFailMessage).End(c, http.StatusUnauthorized)
	c.Abort()
}

func newClaims(entity AuthInterface) claims {
	return claims{
		entity.GetTopic(),
		entity.GetCheckData(),
		jwt.StandardClaims{
			ExpiresAt: entity.ExpiredAt(),
		},
	}
}

func NewToken(entity AuthInterface) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims(entity))
	rs, err := token.SignedString([]byte(config.Config.Application[server.Mode].JwtToken))
	if err != nil {
		return "", err
	}
	return rs, nil
}

func ParseToken(sign string) (*claims, error) {
	token, err := jwt.ParseWithClaims(sign, &claims{}, func(token *jwt.Token) (i interface{}, e error) {
		return []byte(config.Config.Application[server.Mode].JwtToken), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("can't decode token info")
}
