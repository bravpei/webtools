package utils

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

type JwtUtil struct {
}

var jwtUtil JwtUtil

func GetJwtUtil() JwtUtil {
	return jwtUtil
}
func (JwtUtil) Generate(customClaims jwt.Claims, signKey []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
	return token.SignedString(signKey)
}
func (JwtUtil) Parse(tokenStr string, signKey []byte) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return signKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims, nil
}