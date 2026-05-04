package middleware

import (
	"strings"

	viva_api "github.com/Armor-ru/viva-api/internal/app"

	"github.com/Armor-ru/sds-go/pkg/logger"
	httpt "github.com/Armor-ru/sds-go/pkg/transport/http"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// InitExtHTTP вешает CORS на лендинг и подпись X-Signature на остальные внешние POST.
func InitExtHTTP(v *viva_api.Viva) {
	ht, ok := v.ExtTransport.(*httpt.Transport)
	sig := httpt.Signature(httpt.SignatureConfig{Secrets: v.Secrets, Header: "X-Signature"})
	if !ok {
		logger.Warn().Msg("extTransport is not *http.Transport; webhook signature middleware skipped")
		return
	}
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
