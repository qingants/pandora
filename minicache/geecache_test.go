package minicache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetterFunc_Get(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		f       GetterFunc
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"normal test",
			GetterFunc(func(key string) ([]byte, error) {
				return []byte(key), nil
			}),
			args{key: "aabbccdd"},
			[]byte("aabbccdd"),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.f.Get(tt.args.key)
			if (err != nil) == tt.wantErr {
				t.Errorf("GetterFunc.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetterFunc.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGet(t *testing.T) {
	var db = map[string]string{
		"rocky": "handsame",
		"amy":   "love",
		"dim":   "cute",
	}
	loadCounts := make(map[string]int, len(db))

	g := NewGroup("name", 2<<10, GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if v, ok := db[key]; ok {
			if _, ok := loadCounts[key]; ok {
				loadCounts[key] = 0
			}
			loadCounts[key] += 1
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))

	for k, v := range db {
		if view, err := g.Get(k); err != nil || view.String() != v {
			t.Fatalf("failed to get value %s", v)
		}
		if _, err := g.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	unknow := "unknow"
	if view, err := g.Get(unknow); err == nil {
		t.Fatalf("miss key %s, got %s", unknow, view)
	}
}

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Fatalf("callback failed")
	}
}

func TestGetGroup(t *testing.T) {
	groupName := "scores"
	NewGroup(groupName, 2<<10, GetterFunc(func(key string) (bytes []byte, err error) {
		return
	}))
	if group := GetGroup(groupName); group == nil || group.name != groupName {
		t.Fatalf("group name %v not exists", groupName)
	}
	if group := GetGroup(groupName + "xxx"); group != nil {
		t.Fatalf("expect nil, but got %v", group.name)
	}
}
