package minirpc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int
	CodecType   string
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   GobType,
}

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (s *Server) Accept(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("rpc server: accept error ", err)
			return
		}
		go s.ServerConn(conn)
	}
}

func (s *Server) ServerConn(conn io.ReadWriteCloser) {
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
	s.serverCodec(f(conn))
}

var invalidRequest = struct{}{}

func (s *Server) serverCodec(codec Codec) {
	// sending := new(sync.Mutex)
	// wg := new(sync.WaitGroup)
	for {
		req, err := s.readRequest(codec)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			s.sendResponse(codec, req.h, invalidRequest)
			continue
		}
		// wg.Add(1)
		go s.handleRequest(codec, req)
	}
	// wg.Wait()
	err := codec.Close()
	if err != nil {
		log.Printf("codec close error %s\n", err.Error())
	}
}

type request struct {
	h          *Head
	arg, reply reflect.Value
}

func (s *Server) readHead(cc Codec) (*Head, error) {
	var head Head
	if err := cc.ReadHead(&head); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Printf("rpc server: read head err: %s", err.Error())
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
	body := request{
		h: head,
	}
	body.arg = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(body.arg.Interface()); err != nil {
		log.Println("rpc server: read argv err: ", err)
	}

	return &body, nil
}

func (s *Server) handleRequest(cc Codec, r *request) {
	log.Printf("minirpc server seq %d ", r.h.Seq)
	r.reply = reflect.ValueOf(fmt.Sprintf("minirpc resp %d", r.h.Seq))
	s.sendResponse(cc, r.h, r.arg.Interface())
}

func (s *Server) sendResponse(cc Codec, head *Head, body any) {
	if err := cc.Write(head, body); err != nil {
		log.Println("rpc server: write response error ", err)
	}
}

func Accept(l net.Listener) {
	DefaultServer.Accept(l)
}
