package utils

import (
	"encoding/hex"
	"math/rand"
	"time"
)

// GenerateSimpleID 生成一个8位的ID，不进行错误处理
func GenerateSimpleID() string {
	// 设置随机数种子（只需要设置一次）
	// 注意：在实际项目中，通常在程序初始化时设置一次即可
	seed := time.Now().UnixNano()
	localRand := rand.New(rand.NewSource(seed))

	// 生成4字节随机数据（hex编码后变成8位）
	bytes := make([]byte, 4)
	localRand.Read(bytes) // 忽略错误

	// 返回8位十六进制字符串
	return hex.EncodeToString(bytes)
}
