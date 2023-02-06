package lru

import (
	"container/list"
	"log"
	"os"
	"reflect"
	"testing"
)

type Str string

func (d Str) Len() int {
	return len(d)
}

func setup() {
	log.Println("set up")
	c := New(100, func(s string, v Value) { log.Println("xxxxx") })
	log.Println(c)
}

func teardown() {
	log.Println("tear down")
}

func TestMain(m *testing.M) {
	log.Println("testMain")
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func TestCache(t *testing.T) {
	// <setup code>
	t.Helper()
	t.Run("A=1", func(t *testing.T) { t.Log("---->1") })
	t.Run("A=2", func(t *testing.T) { t.Log("---->2") })
	// <teardown code>
}

func TestCache_Get(t *testing.T) {
	type fields struct {
		maxBytes  int64
		nbytes    int64
		ll        *list.List
		cache     map[string]*list.Element
		OnEvicted func(key string, value Value)
	}
	type args struct {
		key string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantValue Value
		wantOk    bool
	}{
		// TODO: Add test cases.
		{
			name: "empty cache",
			fields: fields{
				maxBytes: 100,
				nbytes:   0,
				ll:       list.New(),
				cache:    make(map[string]*list.Element),
				OnEvicted: func(key string, value Value) {
					print(key, value)
				},
			},
			args:      args{key: "rocky"},
			wantValue: nil,
			wantOk:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cache{
				maxBytes:  tt.fields.maxBytes,
				nbytes:    tt.fields.nbytes,
				ll:        tt.fields.ll,
				cache:     tt.fields.cache,
				OnEvicted: tt.fields.OnEvicted,
			}
			gotValue, gotOk := c.Get(tt.args.key)
			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("Cache.Get() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
			if gotOk == tt.wantOk {
				t.Errorf("Cache.Get() gotOk = %v, want %v", gotOk, tt.wantOk)
			}

			c.Add("first", Str("abcd"))
			gotValue, gotOk = c.Get("first")
			if !gotOk || string(gotValue.(Str)) != "abcd" {
				t.Fatalf("cache hit first=abcd failed")
			}
			if _, ok := c.Get("second"); ok {
				t.Fatalf("cache miss second failed")
			}
		})
	}
}

func TestCache_Disuse(t *testing.T) {
	k1, v1, k2, v2, k3, v3 := "k1", "v1", "k2", "v2", "k3", "v3"
	cap := len(k1 + v1 + k2 + v2)
	lru := New(int64(cap), nil)
	lru.Add(k1, Str(v1))
	lru.Add(k2, Str(v2))
	lru.Add(k3, Str(v3))

	if _, ok := lru.Get(k1); ok || lru.Len() != 2 {
		t.Fatalf("disuse k1 failure")
	}
}

func TestCache_Add(t *testing.T) {
	type fields struct {
		maxBytes  int64
		nbytes    int64
		ll        *list.List
		cache     map[string]*list.Element
		OnEvicted func(key string, value Value)
	}
	type args struct {
		key   string
		value Value
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cache{
				maxBytes:  tt.fields.maxBytes,
				nbytes:    tt.fields.nbytes,
				ll:        tt.fields.ll,
				cache:     tt.fields.cache,
				OnEvicted: tt.fields.OnEvicted,
			}
			c.Add(tt.args.key, tt.args.value)
		})
	}
}

func TestCache_Len(t *testing.T) {
	type fields struct {
		maxBytes  int64
		nbytes    int64
		ll        *list.List
		cache     map[string]*list.Element
		OnEvicted func(key string, value Value)
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "len",
			fields: fields{
				ll: list.New(),
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cache{
				maxBytes:  tt.fields.maxBytes,
				nbytes:    tt.fields.nbytes,
				ll:        tt.fields.ll,
				cache:     tt.fields.cache,
				OnEvicted: tt.fields.OnEvicted,
			}
			if got := c.Len(); got != tt.want {
				t.Errorf("Cache.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOnEvited(t *testing.T) {
	keys := make([]string, 0)
	cb := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(10), cb)
	lru.Add("key1", Str("123456"))
	lru.Add("k2", Str("v2"))
	lru.Add("k3", Str("v3"))
	lru.Add("k4", Str("v4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("call onEvited failed, expect keys %s equals to %s", keys, expect)
	}
}
