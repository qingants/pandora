package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/qingants/pandora/miniweb"
)

func middlewareExample() miniweb.HandleFunc {
	return func(ctx *miniweb.Context) {
		t := time.Now()
		ctx.Fail(500, "Internal Server Error")
		log.Printf("[%d] %s in %v for group v2", ctx.StatusCode, ctx.Request.RequestURI, time.Since(t))
	}
}

func main() {
	r := miniweb.NewEngine()
	r.Use(miniweb.Logger())

	// r.GET("/", func(ctx *miniweb.Context) {
	// 	ctx.HTML(http.StatusOK, "<h1>mini<h1>")
	// })
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
		// v1.GET("/", func(ctx *miniweb.Context) {
		// 	ctx.HTML(http.StatusOK, "<h2>hi, mini v1</h2>")
		// })
		v1.GET("/hi", func(ctx *miniweb.Context) {
			ctx.String(http.StatusOK, "hi %s, you are at %s\n", ctx.Query("name"), ctx.Path)
		})
	}

	v2 := r.Group("v2")
	v2.Use(middlewareExample())
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

	FormatAsDate := func(t time.Time) string {
		year, month, day := t.Date()
		return fmt.Sprintf("%d-%02d-%02d", year, month, day)
	}

	r.Static("/assets", "/Users/luotuo/gospace/src/pandora/miniweb/example/static")
	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})

	r.LoadHTMLGlob("/Users/luotuo/gospace/src/pandora/miniweb/example/templates/*")

	r.GET("/tmpl", func(ctx *miniweb.Context) {
		ctx.HTML(http.StatusOK, "mini.tmpl", nil)
	})

	r.Run("127.0.0.1:8888")
}
