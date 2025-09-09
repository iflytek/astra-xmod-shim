package http

import (
	"github.com/gin-gonic/gin"
)

// Server 通用HTTP服务器
type Server struct {
	engine *gin.Engine // 内部维护gin引擎
	addr   string
}

// NewServer 创建HTTP服务器实例
func NewServer(addr string) *Server {
	return &Server{
		engine: gin.Default(),
		addr:   addr,
	}
}

// GetEngine 提供引擎访问方法（修正引用问题的核心）
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}

// Run 启动服务器
func (s *Server) Run() error {
	return s.engine.Run(s.addr)
}
