// utils/common.go
package utils

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// 初始化随机数种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// MaskPhone 手机号脱敏（如：13812345678 → 138****5678）
// phone：需要脱敏的手机号（11位）
// 返回：脱敏后的手机号
func MaskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

// FormatTime 日期时间格式化（默认格式：2006-01-02 15:04:05）
// t：需要格式化的时间
// 返回：格式化后的字符串
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// FormatTimeCustom 自定义日期时间格式化
// t：需要格式化的时间
// layout：自定义格式（如：2006-01-02 / 15:04）
// 返回：格式化后的字符串
func FormatTimeCustom(t time.Time, layout string) string {
	if t.IsZero() || layout == "" {
		return ""
	}
	return t.Format(layout)
}

// GenerateOrderNo 生成唯一订单编号（格式：YYYYMMDDHHMMSS + 6位随机数）
// 返回：唯一订单号字符串
func GenerateOrderNo() string {
	// 时间戳部分（年月日时分秒）
	timeStr := time.Now().Format("20060102150405")
	// 6位随机数
	randomNum := rand.Intn(900000) + 100000 // 生成100000-999999的随机数
	// 拼接订单号
	return timeStr + strconv.Itoa(randomNum)
}

// GenerateSerialNo 生成唯一明细编号（格式：前缀 + YYYYMMDD + 8位随机数）
// prefix：编号前缀（如：INC-收入/WDR-提现）
// 返回：唯一明细编号字符串
func GenerateSerialNo(prefix string) string {
	// 时间戳部分（年月日）
	timeStr := time.Now().Format("20060102")
	// 8位随机数
	randomNum := rand.Intn(90000000) + 10000000 // 生成10000000-99999999的随机数
	// 拼接明细编号
	return prefix + timeStr + strconv.Itoa(randomNum)
}

// KeepTwoDecimal 金额保留两位小数（避免浮点型计算精度问题）
// amount：原始金额
// 返回：保留两位小数后的金额
func KeepTwoDecimal(amount float64) float64 {
	return math.Round(amount*100) / 100
}

// IsEmptyString 判断字符串是否为空（包含纯空格）
// s：需要判断的字符串
// 返回：true-空 / false-非空
func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

// GetRandomString 生成指定长度的随机字符串（字母+数字）
// length：字符串长度
// 返回：随机字符串
func GetRandomString(length int) string {
	// 随机字符池
	chars := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charLen := len(chars)
	var builder strings.Builder

	for i := 0; i < length; i++ {
		// 随机获取字符索引
		index := rand.Intn(charLen)
		builder.WriteByte(chars[index])
	}

	return builder.String()
}
