package reply

import (
	"bytes"
	"strconv"
)

type Reply interface {
	ToBytes() []byte
}

// + 简单字符串
// - 错误
// : 整数
// $ 字符串
// * 数组

const CRLF = "\r\n"

type SimpleStringReply struct {
	Str string
}

func (r *SimpleStringReply) ToBytes() []byte {
	return []byte("+" + r.Str + CRLF)
}

type ErrorReply struct {
	Err string
}

func (r *ErrorReply) ToBytes() []byte {
	return []byte("-" + r.Err + CRLF)
}

type IntegerReply struct {
	Int int64
}

func (r *IntegerReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Int, 10) + CRLF)
}

type BulkReply struct {
	Arg []byte
}

func (r *BulkReply) ToBytes() []byte {
	if len(r.Arg) == 0 {
		return []byte("$-1" + CRLF)
	}
	return []byte("$" + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}

type MultiBulkReply struct {
	Args [][]byte
}

func (r *MultiBulkReply) ToBytes() []byte {
	if len(r.Args) == 0 {
		return []byte("*-1" + CRLF)
	}
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(len(r.Args)) + CRLF)
	for _, v := range r.Args {
		if len(v) == 0 {
			buf.WriteString("$-1" + CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(v)) + CRLF + string(v) + CRLF)
		}
	}
	return buf.Bytes()
}

type EmptyBulkReply struct{}

func (r *EmptyBulkReply) ToBytes() []byte {
	return []byte("$-1\r\n")
}

type EmptyMultiBulkReply struct{}

func (r *EmptyMultiBulkReply) ToBytes() []byte {
	return []byte("*0\r\n")
}
