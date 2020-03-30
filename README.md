### Install
```
go get -u github.com/gofiber/fiber
go get -u github.com/gofiber/csrf
```
### Example
```go
package main

import (
  "github.com/gofiber/fiber"
  "github.com/gofiber/csrf"
)

func main() {
  app := fiber.New()

  app.Get("/", func(c *fiber.Ctx) {
    c.Send(c.Locals("csrf"))
  })

  app.Use(csrf.New())

  app.Post("/register", func(c *fiber.Ctx) {
    c.Send("Welcome!")
  })

  app.Listen(3000)
}
```
