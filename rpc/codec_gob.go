package rpc

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobClientCodec struct {
	rwc    io.ReadWriteCloser
	enc    *gob.Encoder
	dec    *gob.Decoder
	encBuf *bufio.Writer
}

var _ ClientCodec = (*GobClientCodec)(nil)

func NewGobClientCodec(coon io.ReadWriteCloser) ClientCodec {
	encBuf := bufio.NewWriter(coon)
	return &GobClientCodec{
		rwc:    coon,
		enc:    gob.NewEncoder(encBuf),
		dec:    gob.NewDecoder(coon),
		encBuf: encBuf,
	}
}

func (c *GobClientCodec) WriteRequest(r *Request, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		return
	}
	if err = c.enc.Encode(body); err != nil {
		return
	}
	return c.encBuf.Flush()
}

func (c *GobClientCodec) ReadResponseHeader(r *Response) error {
	return c.dec.Decode(r)
}

func (c *GobClientCodec) ReadResponseBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *GobClientCodec) Close() error {
	return c.rwc.Close()
}

type GobServerCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
	closed bool
}

var _ ServerCodec = (*GobServerCodec)(nil)

func init() {
	NewServerCodecFuncMap[GobType] = NewGobServerCodec
}

func NewGobServerCodec(coon io.ReadWriteCloser) ServerCodec {
	encBuf := bufio.NewWriter(coon)
	return &GobServerCodec{
		rwc:    coon,
		enc:    gob.NewEncoder(encBuf),
		dec:    gob.NewDecoder(coon),
		encBuf: encBuf,
	}
}

func (c *GobServerCodec) ReadRequestHeader(r *Request) error {
	return c.dec.Decode(r)
}

func (c *GobServerCodec) ReadRequestBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *GobServerCodec) WriteResponse(r *Response, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		if c.encBuf.Flush() == nil {
			// Gob couldn't encode the header. Should not happen, so if it does,
			// shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding response:", err)
			c.Close()
		}
		return
	}
	if err = c.enc.Encode(body); err != nil {
		if c.encBuf.Flush() == nil {
			// Was a gob problem encoding the body but the header has been written.
			// Shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding body:", err)
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *GobServerCodec) Close() error {
	return c.rwc.Close()
}
