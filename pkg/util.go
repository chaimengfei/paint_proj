package pkg

import (
	"fmt"
	"time"
)

// GenerateOrderNo 生成包含用户ID的订单号
// 格式: 年月日 + 用户ID后4位 + 4位随机数
// 示例: 2023061598761234 (用户ID后4位9876)
func GenerateOrderNo(userID int64) string {
	now := time.Now()
	datePart := now.Format("20060102")
	userPart := fmt.Sprintf("%04d", userID%10000)
	randomPart := fmt.Sprintf("%06d", now.Nanosecond()/100000%10000)
	return OrderPrefix + datePart + userPart + randomPart
}
