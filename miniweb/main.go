package main

import (
	"net/http"

	"github.com/qingants/pandora/miniweb/mini"
)

func main() {
	r := mini.NewEngine()
	r.GET("/", func(ctx *mini.Context) {
		ctx.HTML(http.StatusOK, "<h1>mini<h1>")
	})
	r.GET("/hi", func(ctx *mini.Context) {
		ctx.String(http.StatusOK, "hi")
	})
	r.GET("/hi/:name", func(ctx *mini.Context) {
		ctx.String(http.StatusOK, "hi,%s Path:%s", ctx.Params["name"], ctx.Path)
	})
	r.GET("/assets/*.css", func(ctx *mini.Context) {
		ctx.JSON(http.StatusOK, mini.H{"assets": ctx.Param(".css")})
	})
	r.Run("127.0.0.1:8888")
}
