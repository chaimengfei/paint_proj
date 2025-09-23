package pkg

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWTToken(userID int64, shopID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"shop_id": shopID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JwtSecret))
}

func ParseJWTToken(tokenStr string) (int64, int64, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(JwtSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, 0, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		userID, userIDOk := claims["user_id"].(float64)
		shopID, shopIDOk := claims["shop_id"].(float64)
		if userIDOk && shopIDOk {
			return int64(userID), int64(shopID), nil
		}
	}
	return 0, 0, errors.New("invalid token")
}
