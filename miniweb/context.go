package miniweb

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type H map[string]any

type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	Method     string
	Path       string
	Params     map[string]string
	StatusCode int

	handlers []HandleFunc
	index    int

	engine *Engine
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: r,
		Method:  r.Method,
		Path:    r.URL.Path,
		index:   -1,
	}
}

func (c *Context) Next() {
	for c.index++; c.index < len(c.handlers); c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) PostForm(key string) string {
	return c.Request.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...any) {
	c.SetHeader("Content-type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj any) {
	c.SetHeader("Content-type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) HTML(code int, name string, data any) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	log.Printf("ExecuteTemplate params name=%s data=%v", name, data)
	if err := c.engine.htmlTemplate.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
	// if err := c.engine.htmlTemplate.ExecuteTemplate(c.Writer, "T", "<script>alert('you have been pwned')</script>"); err != nil {
	// 	c.Fail(500, err.Error())
	// }
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) Fail(code int, msg string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": msg})
}
