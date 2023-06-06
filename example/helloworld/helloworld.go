package main

import (
	"fmt"
	"net/http"

	"github.com/tekqer/hr"
)

func main() {
	r := hr.Default()
	r.GET("/ping", func(c *hr.Ctx) error {
		fmt.Fprint(c, "pong")
		return nil
	})
	http.ListenAndServe(":8080", r)
}
