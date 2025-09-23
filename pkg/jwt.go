package pkg

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 小程序 JWT Token 生成
func GenerateJWTToken(userID int64, shopID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"shop_id":    shopID,
		"token_type": "mini_program",
		"exp":        time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JwtSecret))
}

// 后台管理 JWT Token 生成
func GenerateAdminJWTToken(operatorID int64, operatorName string, shopID int64, isRoot bool) (string, error) {
	claims := jwt.MapClaims{
		"operator_id":   operatorID,
		"operator_name": operatorName,
		"shop_id":       shopID,
		"is_root":       isRoot,
		"token_type":    "admin",
		"exp":           time.Now().Add(time.Hour * 2).Unix(), // 后台管理 token 2小时有效
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JwtSecret))
}

// 小程序 JWT Token 解析
func ParseJWTToken(tokenStr string) (int64, int64, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(JwtSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, 0, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		tokenType, _ := claims["token_type"].(string)
		if tokenType != "mini_program" {
			return 0, 0, errors.New("invalid token type")
		}
		userID, userIDOk := claims["user_id"].(float64)
		shopID, shopIDOk := claims["shop_id"].(float64)
		if userIDOk && shopIDOk {
			return int64(userID), int64(shopID), nil
		}
	}
	return 0, 0, errors.New("invalid token")
}

// 后台管理 JWT Token 解析
func ParseAdminJWTToken(tokenStr string) (int64, string, int64, bool, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(JwtSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, "", 0, false, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		tokenType, _ := claims["token_type"].(string)
		if tokenType != "admin" {
			return 0, "", 0, false, errors.New("invalid token type")
		}
		operatorID, operatorIDOk := claims["operator_id"].(float64)
		operatorName, operatorNameOk := claims["operator_name"].(string)
		shopID, shopIDOk := claims["shop_id"].(float64)
		isRoot, isRootOk := claims["is_root"].(bool)
		if operatorIDOk && operatorNameOk && shopIDOk && isRootOk {
			return int64(operatorID), operatorName, int64(shopID), isRoot, nil
		}
	}
	return 0, "", 0, false, errors.New("invalid token")
}
