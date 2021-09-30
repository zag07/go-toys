package rpc

import "io"

type ServerCodec interface {
	ReadRequestHeader(*Request) error
	ReadRequestBody(interface{}) error
	WriteResponse(*Response, interface{}) error

	// Close can be called multiple times and must be idempotent.
	Close() error
}

type Request struct {
	ServiceMethod string   // format: "Service.Method"
	Seq           uint64   // sequence number chosen by client
	next          *Request // for free list in Server
}

type Response struct {
	ServiceMethod string    // echoes that of the Request
	Seq           uint64    // echoes that of the request
	Error         string    // error, if any.
	next          *Response // for free list in Server
}

// A ClientCodec implements writing of RPC requests and
// reading of RPC responses for the client side of an RPC session.
// The client calls WriteRequest to write a request to the connection
// and calls ReadResponseHeader and ReadResponseBody in pairs
// to read responses. The client calls Close when finished with the
// connection. ReadResponseBody may be called with a nil
// argument to force the body of the response to be read and then
// discarded.
// See NewClient's comment for information about concurrent access.
type ClientCodec interface {
	WriteRequest(*Request, interface{}) error
	ReadResponseHeader(*Response) error
	ReadResponseBody(interface{}) error

	Close() error
}

type CodecType string

const GobType CodecType = "application/gob"

type NewServerCodecFunc func(io.ReadWriteCloser) ServerCodec
type NewClientCodecFunc func(io.ReadWriteCloser) ClientCodec

var NewServerCodecFuncMap map[CodecType]NewServerCodecFunc
var NewClientCodecFuncMap map[CodecType]NewClientCodecFunc

func init() {
	NewServerCodecFuncMap = make(map[CodecType]NewServerCodecFunc)
	NewServerCodecFuncMap[GobType] = NewGobServerCodec

	NewClientCodecFuncMap = make(map[CodecType]NewClientCodecFunc)
	NewClientCodecFuncMap[GobType] = NewGobClientCodec
}
