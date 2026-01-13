package tracker

import (
	"bytes"
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/gorilla/websocket"
)

type signaling_protocol_command uint8

const (
	// sent
	sp_authenticate signaling_protocol_command = iota
	sp_subscribe
	sp_unsubscribe
	sp_signal
	sp_challenge
	// server-only messages
	// sent to new users,
	sp_peer
	sp_kick
)

type command_message struct {
	command  signaling_protocol_command
	argument []byte
}

type signaling_connection struct {
	// the cryptographic identity of the user
	lock   sync.Mutex
	id     identity.ID
	conn   *websocket.Conn
	closed atomic.Bool
}

func (signaling_connection *signaling_connection) init() {
	signaling_connection.closed.Store(true)
}

func (signaling_connection *signaling_connection) dial(signaling_url string) (err error) {
	signaling_connection.lock.Lock()
	signaling_connection.conn, _, err = websocket.DefaultDialer.Dial(signaling_url, nil)
	if err == nil {
		signaling_connection.closed.Store(false)
	}
	signaling_connection.lock.Unlock()
	return
}

func (signaling_connection *signaling_connection) accept(rw http.ResponseWriter, r *http.Request) (err error) {
	signaling_connection.lock.Lock()
	signaling_connection.conn, err = upgrader.Upgrade(rw, r, nil)
	if err == nil {
		signaling_connection.closed.Store(false)
	}
	signaling_connection.lock.Unlock()
	return
}

func (signaling_connection *signaling_connection) Closed() bool {
	return signaling_connection.closed.Load()
}

func (signaling_connection *signaling_connection) Close() (err error) {
	if signaling_connection.closed.CompareAndSwap(false, true) {
		err = net.ErrClosed
		return
	}
	signaling_connection.lock.Lock()
	err = signaling_connection.conn.Close()
	signaling_connection.conn = nil
	signaling_connection.lock.Unlock()
	return
}

func (signaling_connection *signaling_connection) command(command signaling_protocol_command, content []byte) {
	if !signaling_connection.closed.Load() {
		var buffer bytes.Buffer
		buffer.WriteByte(byte(command))
		buffer.Write(content)
		signaling_connection.lock.Lock()
		signaling_connection.conn.WriteMessage(websocket.BinaryMessage, buffer.Bytes())
		signaling_connection.lock.Unlock()
	}
}

func (signaling_connection *signaling_connection) receive_command() (command signaling_protocol_command, content []byte, err error) {
	var ws_message_type int
	ws_message_type, content, err = signaling_connection.conn.ReadMessage()
	if err != nil {
		return
	}
	if ws_message_type != websocket.BinaryMessage {
		err = ErrBadCommand
		return
	}
	command = signaling_protocol_command(content[0])
	content = content[1:]
	return
}

type Signal uint8

const (
	Chat Signal = iota
	OfferSDP
	AnswerSDP
	ICECandidate
)
