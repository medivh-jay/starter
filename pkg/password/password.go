package password

import (
	"golang.org/x/crypto/bcrypt"
	"log"
	"starter/pkg/config"
	"starter/pkg/server"
)

var passwordToken = config.Config.Application[server.Mode].PasswordToken

// 密码hash
func Hash(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(passwordToken+password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		return ""
	}

	return string(bytes)
}

// 密码验证
func Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(passwordToken+password))
	return err == nil
}
