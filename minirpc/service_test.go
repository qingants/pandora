package minirpc

import (
	"fmt"
	"reflect"
	"testing"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (F Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

// func (f Foo) sum(args Args, reply *int) error {
// 	*reply = args.Num1 + args.Num2
// 	return nil
// }

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(msg + "\n" + fmt.Sprint(v...))
	}
}

func TestNewService(t *testing.T) {
	var foo Foo
	s := NewService(&foo)
	_assert(len(s.method) == 1, "wrong service method, expect 1, but got %d", len(s.method))
	mType := s.method["Sum"]
	_assert(mType != nil, "wrong service method, expect Sum, but got nil")
}

func TestMethodType(t *testing.T) {
	var foo Foo
	s := NewService(&foo)
	mType := s.method["Sum"]
	argv := mType.newArgv()
	replyv := mType.newReplyv()
	argv.Set(reflect.ValueOf(Args{Num1: 1, Num2: 100}))
	err := s.call(mType, argv, replyv)
	_assert(err == nil && *replyv.Interface().(*int) == 101 && mType.NumCalls() == 1, "fail to call Foo.Sum")
}