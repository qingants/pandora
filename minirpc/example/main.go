package main

import (
	"log"
	"reflect"
	"strings"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	typ := reflect.TypeOf(&wg)

	for ix := 0; ix < typ.NumMethod(); ix++ {
		m := typ.Method(ix)
		argv := make([]string, 0, m.Type.NumIn())
		result := make([]string, 0, m.Type.NumOut())

		for i := 1; i < m.Type.NumIn(); i++ {
			argv = append(argv, m.Type.In(i).Name())
		}

		for i := 0; i < m.Type.NumOut(); i++ {
			result = append(result, m.Type.Out(i).Name())
		}
		log.Printf("func (w *%s) %s(%s) %s",
			typ.Elem().Name(),
			m.Name,
			strings.Join(argv, ","),
			strings.Join(result, ","))
	}
}
