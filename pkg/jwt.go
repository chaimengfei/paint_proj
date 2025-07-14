package pkg

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func GenerateJWTToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JwtSecret))
}

func ParseJWTToken(tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(JwtSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if uid, ok := claims["user_id"].(float64); ok {
			return int64(uid), nil
		}
	}
	return 0, errors.New("invalid token")
}
