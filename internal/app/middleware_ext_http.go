package viva_api

import (
	"strings"

	"github.com/Armor-ru/sds-go/pkg/logger"
	httpt "github.com/Armor-ru/sds-go/pkg/transport/http"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func (s *Viva) initMiddleWare() {
	ht, ok := s.extTransport.(*httpt.Transport)
	sig := httpt.Signature(httpt.SignatureConfig{Secrets: s.secrets, Header: "X-Signature"})
	if !ok {
		logger.Warn().Msg("extTransport is not *http.Transport; webhook signature middleware skipped")
		return
	}
	// Лендинг со статики (другой порт / CDN) — браузер шлёт cross-origin POST.
	ht.Middleware(cors.New(cors.Config{
		Next: func(c *fiber.Ctx) bool {
			return !strings.HasPrefix(c.Path(), "/landing/")
		},
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type,Accept",
	}))
	ht.Middleware(func(c *fiber.Ctx) error {
		if strings.HasPrefix(c.Path(), "/landing/") {
			return c.Next()
		}
		return sig(c)
	})
}
