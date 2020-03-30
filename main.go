// 🚀 Fiber is an Express inspired web framework written in Go with 💖
// 📌 API Documentation: https://fiber.wiki
// 📝 Github Repository: https://github.com/gofiber/fiber
// 🙏 Credits to github.com/labstack/echo/blob/master/middleware/csrf.go

package csrf

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/akyoto/uuid"
	"github.com/gofiber/fiber"
)

// Config ...
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool

	// TokenLength is the length of the generated token.
	TokenLength uint8
	// Optional. Default value 32.

	// TokenLookup is a string in the form of "<source>:<key>" that is used
	// to extract token from the request.
	// Optional. Default value "header:X-CSRF-Token".
	// Possible values:
	// - "header:<name>"
	// - "form:<name>"
	// - "query:<name>"
	TokenLookup string

	// Context key to store generated CSRF token into context.
	// Optional. Default value "csrf".
	ContextKey string

	// Name of the CSRF cookie. This cookie will store CSRF token.
	// Optional. Default value "csrf".
	CookieName string

	// Domain of the CSRF cookie.
	// Optional. Default value none.
	CookieDomain string

	// Path of the CSRF cookie.
	// Optional. Default value none.
	CookiePath string

	// Max age (in seconds) of the CSRF cookie.
	// Optional. Default value 86400 (24hr).
	CookieMaxAge int

	// Indicates if CSRF cookie is secure.
	// Optional. Default value false.
	CookieSecure bool

	// Indicates if CSRF cookie is HTTP only.
	// Optional. Default value false.
	CookieHTTPOnly bool
}

// New ...
func New(config ...Config) func(*fiber.Ctx) {
	// Init config
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.TokenLength == 0 {
		cfg.TokenLength = 32
	}
	if cfg.TokenLookup == "" {
		cfg.TokenLookup = "header:X-CSRF-Token"
	}
	if cfg.ContextKey == "" {
		cfg.ContextKey = "csrf"
	}
	if cfg.CookieName == "" {
		cfg.CookieName = "_csrf"
	}
	if cfg.CookieMaxAge == 0 {
		cfg.CookieMaxAge = 86400
	}
	parts := strings.Split(cfg.TokenLookup, ":")
	extractor := csrfFromHeader(parts[1])
	switch parts[0] {
	case "form":
		extractor = csrfFromForm(parts[1])
	case "query":
		extractor = csrfFromQuery(parts[1])
	case "param":
		extractor = csrfFromParam(parts[1])
	}
	return func(c *fiber.Ctx) {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			c.Next()
			return
		}
		key := c.Cookies(cfg.CookieName)
		token := ""
		if key == "" {
			token = uuid.New().String()
		} else {
			token = key
		}
		switch c.Method() {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			// Validate token only for requests which are not defined as 'safe' by RFC7231
			clientToken, err := extractor(c)
			if err != nil {
				c.SendStatus(fiber.StatusBadRequest)
				return
			}
			if subtle.ConstantTimeCompare([]byte(token), []byte(clientToken)) != 1 {
				c.SendStatus(fiber.StatusForbidden)
				return
			}
		}
		// Set CSRF cookie
		cookie := new(fiber.Cookie)
		cookie.Name = cfg.CookieName
		cookie.Value = token
		if cfg.CookiePath != "" {
			cookie.Path = cfg.CookiePath
		}
		if cfg.CookieDomain != "" {
			cookie.Domain = cfg.CookieDomain
		}
		cookie.Expires = time.Now().Add(time.Duration(cfg.CookieMaxAge) * time.Second)
		cookie.Secure = cfg.CookieSecure
		cookie.HTTPOnly = cfg.CookieHTTPOnly
		c.Cookie(cookie)

		// Store token in context
		c.Locals(cfg.ContextKey, token)

		// Protect clients from caching the response
		c.Vary(fiber.HeaderCookie)

		c.Next()
	}
}

// csrfFromHeader returns a function that extracts token from the request header.
func csrfFromHeader(param string) func(c *fiber.Ctx) (string, error) {
	return func(c *fiber.Ctx) (string, error) {
		token := c.Get(param)
		if token == "" {
			return "", errors.New("missing csrf token in header")
		}
		return token, nil
	}
}

// csrfcsrfFromQuery returns a function that extracts token from the query string.
func csrfFromQuery(param string) func(c *fiber.Ctx) (string, error) {
	return func(c *fiber.Ctx) (string, error) {
		token := c.Query(param)
		if token == "" {
			return "", errors.New("missing csrf token in query string")
		}
		return token, nil
	}
}

// csrfFromParam returns a function that extracts token from the url param string.
func csrfFromParam(param string) func(c *fiber.Ctx) (string, error) {
	return func(c *fiber.Ctx) (string, error) {
		token := c.Params(param)
		if token == "" {
			return "", errors.New("missing csrf token in url parameter")
		}
		return token, nil
	}
}

// csrfFromParam returns a function that extracts token from the url param string.
func csrfFromForm(param string) func(c *fiber.Ctx) (string, error) {
	return func(c *fiber.Ctx) (string, error) {
		token := c.FormValue(param)
		if token == "" {
			return "", errors.New("missing csrf token in form parameter")
		}
		return token, nil
	}
}
