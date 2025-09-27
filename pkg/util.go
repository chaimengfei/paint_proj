package pkg

import (
	"fmt"
	"time"
)

// GenerateOperatorNo 生成包含用户ID的操作单号
// 格式: 年月日  + 4位随机数
func GenerateOrderNo(prefix string, userID int64) string {
	now := time.Now()
	datePart := now.Format("20060102")
	randomPart := fmt.Sprintf("%04d", now.Nanosecond()/100000%10000)
	return prefix + datePart + randomPart
}
