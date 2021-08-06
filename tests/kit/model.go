package kit

import (
	"bufio"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/beyondstorage/beyond-ftp/utils"
)

const (
	_1 = iota + 1
	_2
	_3
	_4
	_5
)

type modelType int

const (
	reply modelType = iota
	waitReply
	restart
	login
	acctOrRnto
	rnfr
	generalized
)

type code int

func (c code) tp() int {
	return int(c / 100)
}

type state int

const (
	begin state = iota
	wait
	erro
	success
	failure
	another
)

type model struct {
	t  *testing.T
	tp modelType

	conn     utils.Conn
	tpMapper func(t code) state

	respMsg []string
}

func (m *model) Begin(cmd string) *model {
	send(bufio.NewWriter(m.conn), cmd)
	return m
}

func (m *model) Expect(states ...state) *model {
	return m.ExpectWithMsg("", states...)
}

func (m *model) ExpectWithMsg(message string, states ...state) *model {
	c, msg := response(bufio.NewReader(m.conn))
	if message != "" {
		assert.Equal(m.t, message, msg)
	}
	m.respMsg = append(m.respMsg, msg)
	curState := m.tpMapper(c)

	anyOf := utils.AnyOf(states, func(i int) bool {
		return states[i] == curState
	})
	assert.True(m.t, anyOf, fmt.Sprintf("expect one of %v, actual %v", states, curState))

	return m
}

func (m *model) TakeAction(f func()) *model {
	f()
	return m
}

func (m *model) Wait(msg ...string) *model {
	m.ExpectWithMsg(first(msg), wait)
	return m
}

func (m *model) Failure(msg ...string) *model {
	m.ExpectWithMsg(first(msg), failure)
	return m
}

func (m *model) Success(msg ...string) *model {
	m.ExpectWithMsg(first(msg), success)
	return m
}

func (m *model) Error(msg ...string) *model {
	m.ExpectWithMsg(first(msg), erro)
	return m
}

func (m *model) Another() *model {
	m.Expect(another)
	return m
}

func (m *model) Any() *model {
	m.Expect(erro, failure, success)
	return m
}

func (m *model) Auto() *model {
	switch m.tp {
	case reply, login, rnfr, acctOrRnto:
		return m
	case waitReply:
		return m.Expect(wait)
	}

	panic("no reachable")
}

func (m *model) message() []string {
	return m.respMsg
}

func first(msg []string) string {
	if len(msg) == 0 {
		return ""
	}
	return msg[0]
}

func newModel(t *testing.T, tp modelType, conn utils.Conn, mapper func(t code) state) *model {
	return &model{
		t:        t,
		tp:       tp,
		conn:     conn,
		tpMapper: mapper,
	}
}

func replyModel(t *testing.T, conn utils.Conn) *model {
	return newModel(t, reply, conn, func(t code) state {
		switch t.tp() {
		case _1, _3:
			return erro
		case _2:
			return success
		case _4, _5:
			return failure
		}
		return erro
	})
}

func waitReplyModel(t *testing.T, conn utils.Conn) *model {
	return newModel(t, waitReply, conn, func(t code) state {
		switch t.tp() {
		case _1:
			return wait
		case _3:
			return erro
		case _2:
			return success
		case _4, _5:
			return failure
		}
		return erro
	})
}

func loginModel(t *testing.T, conn utils.Conn) *model {
	return newModel(t, login, conn, func(t code) state {
		switch t.tp() {
		case _1:
			return erro
		case _2:
			return success
		case _3:
			return another
		case _4, _5:
			return failure
		}
		return erro
	})
}

func acctOrRntoModel(t *testing.T, conn utils.Conn) *model {
	return newModel(t, acctOrRnto, conn, func(t code) state {
		switch t.tp() {
		case _1, _3:
			return erro
		case _2:
			return success
		case _4, _5:
			return failure
		}
		return erro
	})
}

func rnfrModel(t *testing.T, conn utils.Conn) *model {
	return newModel(t, rnfr, conn, func(t code) state {
		switch t.tp() {
		case _1, _2:
			return erro
		case _3:
			return another
		case _4, _5:
			return failure
		}
		return erro
	})
}
