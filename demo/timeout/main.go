package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	result := 20.4 * 100
	fmt.Println(result) // 输出 2040.0000000000002

}

func withTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		cancel = func() {
			fmt.Println("time out")
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func longTimeoutHandler(c *gin.Context) {
	// 模拟长时间操作
	time.Sleep(20 * time.Second)

	c.String(http.StatusOK, "Long Timeout Done")
}

func shortTimeoutHandler(c *gin.Context) {
	// 模拟短时间操作
	time.Sleep(1 * time.Second)

	c.String(http.StatusOK, "Short Timeout Done")
}
