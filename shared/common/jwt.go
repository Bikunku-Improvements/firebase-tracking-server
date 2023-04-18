package common

import (
	"fmt"
	"time"
	"tracking-server/shared/config"

	"github.com/golang-jwt/jwt"
)

func NewJWT(username string, env *config.EnvConfig) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * 8).Unix(),
		IssuedAt:  time.Now().Unix(),
		Issuer:    username,
	})

	return token.SignedString([]byte(env.JWTSecret))
}

func NewJWTWithID(username string, id string, env *config.EnvConfig) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * 8).Unix(),
		Id: id,
		IssuedAt:  time.Now().Unix(),
		Issuer:    username,
	})

	return token.SignedString([]byte(env.JWTSecret))
}

func parseJWT(tokenString string, env *config.EnvConfig) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(env.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func ExtractTokenData(tokenString string, env *config.EnvConfig) (string, string, error) {
	token, err := parseJWT(tokenString, env)
	if err != nil {
		return "", "", err
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	return claims["iss"].(string), claims["jti"].(string), nil
}
