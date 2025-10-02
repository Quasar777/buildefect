package middleware

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// get token from request headers
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing Authorization header"})
		}
		// split "Bearer" and token itself
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid Authorization header"})
		}
		tokenStr := parts[1]

		// parse token
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			// validate signing method if needed
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token claims"})
		}

		// read sub (user id)
		sub, ok := claims["sub"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token subject"})
		}
		uid, err := strconv.ParseUint(sub, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token subject"})
		}

		// set locals for handlers
		c.Locals("user_id", uint(uid))
		if role, ok := claims["role"].(string); ok {
			c.Locals("role", role)
		}
		return c.Next()
	}
}

func RequireRole(role string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        r := c.Locals("role") // получаем роль из JWT
		fmt.Println(r)
        if r == nil || r.(string) != role {
            return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
                "error": "insufficient permissions",
            })
        }
        return c.Next()
    }
}