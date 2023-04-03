package minirpc

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

type methodType struct {
	method    reflect.Method
	argType   reflect.Type
	replyType reflect.Type
	numCalls  uint64
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	if m.argType.Kind() == reflect.Ptr {
		argv = reflect.New(m.argType.Elem())
	} else {
		argv = reflect.New(m.argType).Elem()
	}
	return argv
}

func (m *methodType) newReplyv() reflect.Value {
	replyv := reflect.New(m.replyType.Elem())
	switch m.replyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.replyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.replyType.Elem(), 0, 0))
	}

	return replyv
}

type service struct {
	name   string
	typ    reflect.Type
	val    reflect.Value
	method map[string]*methodType
}

func NewService(val any) *service {
	s := service{
		typ:    reflect.TypeOf(val),
		val:    reflect.ValueOf(val),
		method: make(map[string]*methodType),
	}
	s.name = reflect.Indirect(s.val).Type().Name()
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not a valid service name", s.name)
	}
	s.registerMethods()
	return &s
}

func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argvType, replyType := mType.In(1), mType.In(2)
		if !IsExportedOrBuiltinType(argvType) || !IsExportedOrBuiltinType(replyType) {
			continue
		}

		s.method[method.Name] = &methodType{
			method:    method,
			argType:   mType.In(1),
			replyType: mType.In(2),
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

func IsExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

func (s *service) call(m *methodType, argv, reply reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	result := f.Call([]reflect.Value{s.val, argv, reply})
	if err := result[0].Interface(); err != nil {
		return err.(error)
	}
	return nil
}
