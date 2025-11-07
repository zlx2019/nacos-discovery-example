package server

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"log"
	"strings"
)

var (
	httpServer *fiber.App
)

// StartupHttpServer 启动 HTTP 服务
func StartupHttpServer(host string, port uint64) error {
	httpServer = fiber.New()
	httpServer.Get("/users", func(ctx fiber.Ctx) error {
		name := ctx.Query("name", "default")
		log.Printf("http call: %s \n", name)
		return ctx.SendString(strings.ToUpper(name))
	})
	return httpServer.Listen(fmt.Sprintf("%s:%d", host, port))
}

// Shutdown 关闭 HTTP 服务
func Shutdown(ctx context.Context) error {
	if httpServer != nil {
		return httpServer.ShutdownWithContext(ctx)
	}
	return nil
}
