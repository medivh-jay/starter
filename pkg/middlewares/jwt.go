package middlewares

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"reflect"
	"starter/pkg/app"
	"starter/pkg/config"
	"starter/pkg/mongo"
	"starter/pkg/server"
)

type claims struct {
	Id      interface{} // 唯一id
	LoginAt int64       // 登录时间戳
	jwt.StandardClaims
}

// 认证表信息,该表需要为mongo
type Auth struct {
	Entity  app.Table // 非指针对象
	ParseId func(string) interface{}
}

var AuthInfo Auth
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
			var entity = reflect.New(reflect.TypeOf(AuthInfo.Entity))
			err = mongo.Collection(AuthInfo.Entity).Where(bson.M{"_id": claims.Id}).FindOne(entity.Interface())
			loginAt := entity.Elem().FieldByName("LoginAt")
			if err == nil && loginAt.IsValid() {
				if loginAt.MapIndex(reflect.ValueOf(c.DefaultQuery("platform", "web"))).Interface() != claims.LoginAt {
					app.NewResponse(app.AuthFail, nil, app.AuthFailMessage).End(c, http.StatusUnauthorized)
					c.Abort()
					return
				} else {
					c.Set(AuthKey, entity.Interface())
					c.Header(JwtHeaderKey, token)
					c.Next()
					return
				}
			}
		}
	}
	app.NewResponse(app.AuthFail, nil, app.AuthFailMessage).End(c, http.StatusUnauthorized)
	c.Abort()
}

func newClaims(id interface{}, loginAt, expiresAt int64) claims {
	return claims{
		id,
		loginAt,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}
}

func NewToken(id interface{}, loginAt, expire int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims(id, loginAt, expire))
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
		claims.Id = AuthInfo.ParseId(claims.Id.(string))
		return claims, nil
	}

	return nil, errors.New("can't decode token info")
}
