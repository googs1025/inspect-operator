package common

import (
	"encoding/base64"
)

// EncodeScript base64 转换 string
func EncodeScript(str string) string {
	// 将字符串转换为字节数组
	data := []byte(str)
	return base64.StdEncoding.EncodeToString(data)
}
