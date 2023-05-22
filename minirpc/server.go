package minirpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const MagicNumber = 0x3bef5c

const (
	connected        = "200 connected to mini RPC"
	defaultRPCPath   = "/_minirpc_"
	defaultDebugPath = "/debug/minirpc"
)

type Option struct {
	MagicNumber    int
	CodecType      string
	ConnectTimeout time.Duration
	HandleTimeout  time.Duration
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      GobType,
	ConnectTimeout: time.Second * 10,
}

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (s *Server) Accept(l net.Listener) {
	for {
		log.Printf("accept tcp connection %s", l.Addr().String())
		conn, err := l.Accept()
		if err != nil {
			log.Println("rpc server: accept error ", err)
			return
		}
		go s.ServerConn(conn)
	}
}

func (s *Server) ServerConn(conn io.ReadWriteCloser) {
	log.Printf("%s rpc server accept conn:", reflect.TypeOf(s).String())

	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println("rpc server: close conn error ", err)
		}
	}()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Printf("rpc server: options error %s", err.Error())
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
	}
	f := NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc sever: invalid codec type %s", opt.CodecType)
		return
	}
	s.serveCodec(f(conn), &opt)
}

var invalidRequest = struct{}{}

func (s *Server) serveCodec(codec Codec, opt *Option) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	log.Println("rpc server server codec")
	for {
		req, err := s.readRequest(codec)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			s.sendResponse(codec, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go s.handleRequest(codec, req, sending, wg, opt.HandleTimeout)
	}
	wg.Wait()
	err := codec.Close()
	if err != nil {
		log.Printf("codec close error %s\n", err.Error())
	}
}

type request struct {
	h          *Head
	arg, reply reflect.Value
	mtype      *methodType
	svc        *service
}

func (s *Server) readHead(cc Codec) (*Head, error) {
	log.Println("rpc server.readHead")
	var head Head
	if err := cc.ReadHead(&head); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			debug.PrintStack()
			log.Printf("rpc server: read head error: %s", err.Error())
		}
		return nil, err
	}
	return &head, nil
}

func (s *Server) readRequest(cc Codec) (*request, error) {
	head, err := s.readHead(cc)
	if err != nil {
		return nil, err
	}
	req := request{
		h: head,
	}
	req.svc, req.mtype, err = s.findService(head.Method)
	if err != nil {
		return &req, err
	}
	req.arg = req.mtype.newArgv()
	req.reply = req.mtype.newReplyv()
	argvi := req.arg.Interface()
	if req.arg.Type().Kind() != reflect.Ptr {
		argvi = req.arg.Addr().Interface()
	}
	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read argv err: ", err)
		return &req, nil
	}

	return &req, nil
}

func (s *Server) handleRequest(cc Codec, r *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	log.Printf("-- minirpc server seq %d ", r.h.Seq)
	defer wg.Done()

	called := make(chan struct{})
	sent := make(chan struct{})
	go func() {
		err := r.svc.call(r.mtype, r.arg, r.reply)
		called <- struct{}{}
		if err != nil {
			r.h.Error = err.Error()
			s.sendResponse(cc, r.h, invalidRequest, sending)
			sent <- struct{}{}
			return
		}
		s.sendResponse(cc, r.h, r.reply.Interface(), sending)
	}()

	if timeout == 0 {
		<-called
		<-sent
		return
	}

	select {
	case <-time.After(timeout):
		r.h.Error = fmt.Sprintf("rpc server: request handle timeout: expect within %s", timeout)
		s.sendResponse(cc, r.h, invalidRequest, sending)
	case <-called:
		<-sent
	}
}

func (s *Server) sendResponse(cc Codec, head *Head, body any, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()

	if err := cc.Write(head, body); err != nil {
		log.Println("rpc server: write response error ", err)
	}
}

func Accept(l net.Listener) {
	DefaultServer.Accept(l)
}

func (s *Server) Register(rcvr any) error {
	service := NewService(rcvr)
	if _, dup := s.serviceMap.LoadOrStore(service.name, service); dup {
		return errors.New("rpc: service already defined: " + service.name)
	}
	return nil
}

// Register publishes the ceiver's methods in the DefaultServer
func Register(val any) error {
	return DefaultServer.Register(val)
}

func (s *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service.method request ill-formed" + serviceMethod)
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := s.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can not find service" + serviceName)
		return
	}
	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc server: can not find method " + methodName)
	}
	return
}

// ServeHTTP implements an http.Handler that answers RPC request
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = io.WriteString(w, "405 must CONNECT\n")
		return
	}

	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking", r.RemoteAddr, ":", err.Error())
		return
	}

	_, _ = io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	s.ServerConn(conn)
}

// HandleHTTP register an HTTP handler for RPC messages on rpcPath
// and a debugging handler on debug path
// It is still necessary to invoke http.Serve(), typically in a go statement
func (s *Server) HandleHTTP() {
	http.Handle(defaultDebugPath, debugHTTP{s})
	http.Handle(defaultRPCPath, s)
	log.Println("rpc server debug path: ", defaultDebugPath)
}

// HandleHTTP is a convenient approach for default server to register HTTP handlers
func HandleHTTP() {
	DefaultServer.HandleHTTP()
}
