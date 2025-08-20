package pkg

import (
	"fmt"
	"time"
)

// GenerateOperatorNo 生成包含用户ID的操作单号
// 格式: 年月日 + 用户ID后4位 + 4位随机数
// 示例: 2023061598761234 (用户ID后4位9876)
func GenerateOrderNo(prefix string, userID int64) string {
	now := time.Now()
	datePart := now.Format("20060102")
	userPart := fmt.Sprintf("%04d", userID%10000)
	randomPart := fmt.Sprintf("%06d", now.Nanosecond()/100000%10000)
	return prefix + datePart + userPart + randomPart
}
