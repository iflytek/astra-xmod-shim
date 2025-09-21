package utils

import (
	"encoding/hex"
	"math/rand"
	"regexp"
	"strings"
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

// ModelNameToDeploymentName 将模型名转换为 Kubernetes 兼容的 deployment 名称
func ModelNameToDeploymentName(modelName string) string {
	// 1. 转小写
	name := strings.ToLower(modelName)

	// 2. 将 . 替换为 -（或其他策略，如去除）
	// 例如：0.6 → 0-6
	name = strings.ReplaceAll(name, ".", "-")

	// 3. 只保留：字母、数字、-
	// 使用正则替换非法字符
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	name = reg.ReplaceAllString(name, "")

	// 4. 确保以字母开头
	if len(name) == 0 || !isLetter(name[0]) {
		name = "model-" + name
	}

	// 5. 确保以字母或数字结尾
	for len(name) > 0 && !isAlnum(name[len(name)-1]) {
		name = name[:len(name)-1]
	}

	// 6. 限制长度（63字符）
	if len(name) > 63 {
		name = name[:63]
	}

	// 7. 防止全连字符或空字符串
	if name == "" {
		name = "model-default"
	}
	for strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		name = strings.Trim(name, "-")
	}
	if name == "" {
		name = "model"
	}

	return name
}

// 工具函数
func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z')
}

func isAlnum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}
