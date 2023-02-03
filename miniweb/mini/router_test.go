package mini

import (
	"fmt"
	"reflect"
	"testing"
)

func NewTestRouter() *router {
	r := newRouter()
	r.addRoute("GET", "/", nil)
	r.addRoute("GET", "/index", nil)
	r.addRoute("GET", "/hi/:name", nil)
	r.addRoute("GET", "/hi/a/b", nil)
	r.addRoute("GET", "/hi/a/b/c", nil)
	r.addRoute("GET", "/hello/:age", nil)
	r.addRoute("GET", "/assets/*.css", nil)
	return r
}

func TestParsePattern(t *testing.T) {
	if !reflect.DeepEqual(parsePattern("/hi/:name"), []string{"hi", ":name"}) {
		t.Fatalf("parse pattern failed")
	}
	if !reflect.DeepEqual(parsePattern("/hi/*"), []string{"hi", "*"}) {
		t.Fatalf("parse pattern error")
	}
	if !reflect.DeepEqual(parsePattern("/hi/*a/*"), []string{"hi", "*a"}) {
		t.Fatalf("parse pattern error")
	}
}

func TestGetRouter(t *testing.T) {
	r := NewTestRouter()
	n, ps := r.getRouter("GET", "/hi/mini")
	if n == nil {
		t.Fatalf("need not nil")
	}
	if n.pattern != "/hi/:name" {
		t.Fatalf("should match /hi/:name")
	}
	if ps["name"] != "mini" {
		t.Fatalf("name should be 'mini'")
	}
	t.Logf("match path: %s, params['name']: %s\n", n.pattern, ps["name"])

	n, ps = r.getRouter("GET", "/assets/style.css")
	if n.pattern != "/assets/*.css" || ps[".css"] != "style.css" {
		t.Fatalf("pattern should be /assets/*.css && .css should be style.css")
	}

	n, ps = r.getRouter("GET", "/assets/api.js")
	if n.pattern != "/assets/*.css" && ps[".css"] != "api.js" {
		t.Fatalf("pattern should be /assets/*.css & .css should be style.css")
	}
}

func TestGetRoutes(t *testing.T) {
	r := NewTestRouter()
	nodes := r.getRouters("GET")
	for i, n := range nodes {
		fmt.Println(i+1, n)
	}
	if len(nodes) != 7 {
		t.Fatalf("the number of routers should be 5")
	}
}
