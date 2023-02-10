package minirpc

import (
	"bufio"
	"encoding/gob"
	"io"
)

type GobCodec struct {
	conn    io.ReadWriteCloser
	buf     *bufio.Writer
	decoder *gob.Decoder
	encoder *gob.Encoder
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn:    conn,
		buf:     buf,
		decoder: gob.NewDecoder(conn),
		encoder: gob.NewEncoder(buf),
	}
}

func (c *GobCodec) ReadHead(h *Head) error {
	return c.decoder.Decode(h)
}

func (c *GobCodec) ReadBody(body any) error {
	return c.decoder.Decode(body)
}

func (c *GobCodec) Write(h *Head, body any) (err error) {
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()
	if err = c.encoder.Encode(h); err != nil {
		return err
	}
	if err = c.encoder.Encode(body); err != nil {
		return err
	}
	return nil
}

func (c *GobCodec) Close() error {
	return c.conn.Close()
}

var _ Codec = (*GobCodec)(nil)
