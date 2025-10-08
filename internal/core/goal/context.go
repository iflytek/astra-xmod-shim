package goal

import (
	"astron-xmod-shim/internal/core/shimlet"
	dto "astron-xmod-shim/internal/dto/deploy"
)

type Context struct {
	Data       map[string]any // 存储键值对，比如 app_id, url 等
	DeploySpec *dto.DeploySpec
	Shimlet    shimlet.Shimlet
}

// NewContext 创建一个新的上下文实例
func NewContext() *Context {
	return &Context{

		Data: make(map[string]any), // 初始化 map
	}
}

// Set 向上下文中存入一个值
func (c *Context) Set(key string, value any) {
	c.Data[key] = value
}

// Get 从上下文中取出一个值（返回 any，需类型断言）
func (c *Context) Get(key string) any {
	return c.Data[key]
}

// GetString 安全地获取字符串值，如果不存在或类型不对，返回空字符串
func (c *Context) GetString(key string) string {
	if v, ok := c.Data[key].(string); ok {
		return v
	}
	return ""
}
