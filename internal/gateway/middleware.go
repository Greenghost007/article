package gateway

import (
	"article/pkg/logger"
	"article/pkg/tools"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type Middleware struct {
	*BaseApi
}

func NewMiddleware(base *BaseApi) *Middleware {
	return &Middleware{
		base,
	}
}

func (m *Middleware) Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func (m *Middleware) LoggerToFile() gin.HandlerFunc {
	l := logger.LoggerInit(m.conf)
	return func(c *gin.Context) {
		clientIp := c.ClientIP()
		reqMethod := c.Request.Method
		reqURI := c.Request.RequestURI
		statusCode := c.Writer.Status()

		l.Infof("| %s | %s | %s | %d ",
			clientIp,
			reqMethod,
			reqURI,
			statusCode,
		)
	}
}

func (m *Middleware) Auth(source, action string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")

		if tokenString == "" || !strings.HasPrefix(tokenString, "Basic ") {
			tools.BadRequest(ctx, "请登录")
			ctx.Abort()
			return
		}

		tokenString = tokenString[6:]
		uid, err := tools.ParseToken(tokenString, m.conf)
		if err != nil {
			tools.BadRequest(ctx, err.Error())
			ctx.Abort()
			return
		}

		if source != "" && action != "" {
			if err := m.enforcer.Enforce(uid, source, action); err != nil {
				tools.BadRequest(ctx, err.Error())
				ctx.Abort()
				return
			}
		}

		ctx.Set("user_id", uid)
	}
}
