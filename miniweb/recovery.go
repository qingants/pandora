package miniweb

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func traceback(msg string) string {
	var callStack [32]uintptr

	n := runtime.Callers(3, callStack[:])
	var buf strings.Builder
	buf.WriteString(msg + "\nTraceback:\n")
	for _, p := range callStack[:n] {
		filename := runtime.FuncForPC(p)
		file, line := filename.FileLine(p)
		buf.WriteString(fmt.Sprintf("\n\t%s: %d", file, line))
	}
	return buf.String()
}

func Recovery() HandleFunc {
	return func(ctx *Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("%s\n\n", traceback(fmt.Sprintf("%s", r)))
				ctx.Fail(http.StatusInternalServerError, "internal server error")
			}
		}()
		ctx.Next()
	}
}
