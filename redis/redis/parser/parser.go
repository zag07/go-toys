package parser

import (
	"bufio"
	"bytes"
	"errors"
	"go-toys/redis/redis/reply"
	"io"
	"strconv"
	"strings"
)

type Payload struct {
	Data reply.Reply
	Err  error
}

func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse(reader, ch)
	return ch
}

func ParseOne(data []byte) (reply.Reply, error) {
	ch := make(chan *Payload)
	reader := bytes.NewReader(data)
	go parse(reader, ch)
	payload := <-ch
	if payload == nil {
		return nil, errors.New("no reply")
	}
	return payload.Data, payload.Err
}

func ParseBytes(data []byte) ([]reply.Reply, error) {
	ch := make(chan *Payload)
	reader := bytes.NewReader(data)
	go parse(reader, ch)
	var results []reply.Reply
	for payload := range ch {
		if payload == nil {
			return nil, errors.New("no reply")
		}
		if payload.Err!=nil {
			if payload.Err == io.EOF {
				break
			}
			return nil, payload.Err
		}
		results = append(results, payload.Data)
	}
	return results, nil
}

func parse(reader io.Reader, ch chan<- *Payload) {
	bufReader := bufio.NewReader(reader)
	var state readState
	for {
		msg, ioErr, err := readLine(bufReader, &state)
		if err != nil {
			if ioErr {
				ch <- &Payload{Err: err}
				close(ch)
				return
			}
			ch <- &Payload{Err: err}
			state = readState{}
			continue
		}

		if !state.readingMultiLine {
			if msg[0] == '*' {
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{Err: errors.New("parse multi bulk header error")}
					state = readState{}
					continue
				}
				if state.expectedArgsCount == 0 {
					ch <- &Payload{Data: &reply.EmptyMultiBulkReply{}}
					state = readState{}
					continue
				}
			} else if msg[0] == '$' {
				err = parseBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{Err: errors.New("parse bulk header error")}
					state = readState{}
					continue
				}
				if state.bulklen == -1 {
					ch <- &Payload{Data: &reply.EmptyBulkReply{}}
					state = readState{}
					continue
				}
			} else {
				result, err := parseSingleLineReply(msg)
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
				continue
			}
		} else {
			err = readBody(msg, &state)
			if err != nil {
				ch <- &Payload{Err: errors.New("read body error:" + string(msg))}
				state = readState{}
				continue
			}
			if state.finished() {
				var result reply.Reply
				if state.msgType == '*' {
					result = &reply.MultiBulkReply{Args: state.args}
				} else if state.msgType == '$' {
					result = &reply.BulkReply{Arg: state.args[0]}
				}
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
			}
		}
	}

}

func readLine(bufReader *bufio.Reader, state *readState) (msg []byte, ioErr bool, err error) {
	if state.bulklen == 0 {
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
	} else {
		msg = make([]byte, state.bulklen+2)
		_, err = io.ReadFull(bufReader, msg)
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		state.bulklen = 0
	}
	return msg, false, nil
}

func parseMultiBulkHeader(msg []byte, state *readState) error {
	expectedLine, err := strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		// first line of multi bulk reply
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = int(expectedLine)
		state.args = make([][]byte, 0, expectedLine)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

func parseBulkHeader(msg []byte, state *readState) (err error) {
	state.bulklen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if state.bulklen == -1 {
		return nil
	} else if state.bulklen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

func parseSingleLineReply(msg []byte) (result reply.Reply, err error) {
	str := strings.TrimSuffix(string(msg), "\r\n")
	switch msg[0] {
	case '+':
		result = &reply.SimpleStringReply{Str: str[1:]}
	case '-':
		result = &reply.ErrorReply{Err: str[1:]}
	case ':':
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			err = errors.New("protocol error: " + string(msg))
		}
		result = &reply.IntegerReply{Int: val}
	case '$':
		result = &reply.BulkReply{Arg: []byte(str[1:])}
	case '*':
		strs := strings.Split(str, " ")
		args := make([][]byte, len(strs))
		for i, s := range strs {
			args[i] = []byte(s)
		}
		result = &reply.MultiBulkReply{Args: args}
	default:
		err = errors.New("protocol error: " + string(msg))
	}
	return
}

func readBody(msg []byte, state *readState) (err error) {
	line := msg[0 : len(msg)-2]
	if line[0] == '$' {
		state.bulklen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		if state.bulklen <= 0 {
			state.args = append(state.args, []byte{})
			state.bulklen = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return
}

type readState struct {
	readingMultiLine  bool
	expectedArgsCount int
	msgType           byte
	args              [][]byte
	bulklen           int64
}

func (s *readState) finished() bool {
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}
