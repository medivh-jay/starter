package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"net/http"
	"starter/pkg/app"
	"starter/pkg/database"
	"starter/pkg/server"
)

// Claims 生成token的结构体
type Claims struct {
	ID        interface{} // 唯一id
	CheckData string      // 验证信息
	jwt.StandardClaims
}

//// 认证表信息,该表需要为mongo
//type Auth struct {
//	Entity  app.Table // 非指针对象
//	ParseId func(string) interface{}
//}

// AuthInterface 参与 jwt 数据表结构体需要实现这些接口
type AuthInterface interface {
	database.Table
	GetTopic() interface{}                         // 返回唯一信息
	FindByTopic(topic interface{}) AuthInterface   // 根据唯一信息标识获取数据信息, 比如根据用户id获取用户信息,需要注意传入的数据类型
	GetCheckData() string                          // 获取验证信息, jwt加密时, 改信息会一起进行加密, 解密时会解出来然后调用 Check 验证该信息的正确性, 如果是其他数据类型直接转string，比如是个结构体或者map, 直接转为json string
	Check(ctx *gin.Context, checkData string) bool // 验证信息
	ExpiredAt() int64                              // 返回过期时间,时间戳
}

// AuthEntity  具体的jwt验证支持结构体, 需要在自己的应用中赋值指定
var AuthEntity AuthInterface
var (
	// AuthKey 在整个gin.Context 上线文中的 Get 操作的key名,可以获得 AuthEntity
	AuthKey = "users"
	// JwtHeaderKey jwt token 在HTTP请求中的header名
	JwtHeaderKey = "JWT"
)

// VerifyAuth 验证用户有效性的中间件
func VerifyAuth(c *gin.Context) {
	token := c.GetHeader(JwtHeaderKey)
	if token != "" {
		claims, err := ParseToken(token)
		if err == nil {
			var entity = AuthEntity.FindByTopic(claims.ID)
			if entity.Check(c, claims.CheckData) {
				c.Set(AuthKey, entity) // 向下设置用户信息,控制器可直接获取
				c.Header(JwtHeaderKey, token)
				c.Next()
				return
			}
			app.NewResponse(app.AuthFail, nil, app.AuthFailMessage).End(c, http.StatusUnauthorized)
			c.Abort()
			return
		}
	}
	app.NewResponse(app.AuthFail, nil, app.AuthFailMessage).End(c, http.StatusUnauthorized)
	c.Abort()
}

func newClaims(entity AuthInterface) Claims {
	return Claims{
		entity.GetTopic(),
		entity.GetCheckData(),
		jwt.StandardClaims{
			ExpiresAt: entity.ExpiredAt(),
		},
	}
}

// NewToken 根据传入的结构体(非空结构体)返回一个token
func NewToken(entity AuthInterface) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims(entity))
	rs, err := token.SignedString([]byte(server.Config.JwtToken))
	if err != nil {
		return "", err
	}
	return rs, nil
}

// ParseToken 根据传入 token 得到 Claims 信息
func ParseToken(sign string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(sign, &Claims{}, func(token *jwt.Token) (i interface{}, e error) {
		return []byte(server.Config.JwtToken), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("can't decode token info")
}
