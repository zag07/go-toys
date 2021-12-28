package main

import (
	"net/http"
	"testing"
)

func TestWeb(t *testing.T) {
	r := Default()
	r.GET("/", func(c *Context) {
		c.String(http.StatusOK, "Hello Web\n")
	})
	r.GET("/panic", func(c *Context) {
		names := []string{"web"}
		c.String(http.StatusOK, names[100])
	})

	r.Run(":9999")
}