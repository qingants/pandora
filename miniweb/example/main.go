package main

import (
	"net/http"

	"github.com/qingants/pandora/miniweb"
)

func main() {
	r := miniweb.NewEngine()
	r.GET("/", func(ctx *miniweb.Context) {
		ctx.HTML(http.StatusOK, "<h1>mini<h1>")
	})
	r.GET("/hi", func(ctx *miniweb.Context) {
		ctx.String(http.StatusOK, "hi")
	})
	r.GET("/hi/:name", func(ctx *miniweb.Context) {
		ctx.String(http.StatusOK, "hi,%s Path:%s", ctx.Params["name"], ctx.Path)
	})
	r.GET("/assets/*.css", func(ctx *miniweb.Context) {
		ctx.JSON(http.StatusOK, miniweb.H{"assets": ctx.Param(".css")})
	})

	v1 := r.Group("v1")
	{
		v1.GET("/", func(ctx *miniweb.Context) {
			ctx.HTML(http.StatusOK, "<h2>hi, mini v1</h2>")
		})
		v1.GET("/hi", func(ctx *miniweb.Context) {
			ctx.String(http.StatusOK, "hi %s, you are at %s\n", ctx.Query("name"), ctx.Path)
		})
	}

	v2 := r.Group("v2")
	{
		v2.GET("/hi/:name", func(ctx *miniweb.Context) {
			ctx.String(http.StatusOK, "hi %s, you are at %s\n", ctx.Param("name"), ctx.Path)
		})
	}

	auth := v2.Group("auth")
	{
		auth.POST("/sign", func(ctx *miniweb.Context) {
			ctx.JSON(http.StatusOK, miniweb.H{
				"username": ctx.PostForm("username"),
				"password": ctx.PostForm("password"),
			})
		})
	}

	r.Run("127.0.0.1:8888")
}
