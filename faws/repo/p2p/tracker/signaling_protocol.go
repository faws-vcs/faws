package tracker

import (
	"bytes"

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
	id   identity.ID
	conn *websocket.Conn
}

func (signaling_connection *signaling_connection) Open() bool {
	return signaling_connection.conn != nil
}

func (signaling_connection *signaling_connection) Close() (err error) {
	if signaling_connection.conn == nil {
		return
	}
	err = signaling_connection.conn.Close()
	signaling_connection.conn = nil
	return
}

func (signaling_connection *signaling_connection) command(command signaling_protocol_command, content []byte) {
	var buffer bytes.Buffer
	buffer.WriteByte(byte(command))
	buffer.Write(content)
	signaling_connection.conn.WriteMessage(websocket.BinaryMessage, buffer.Bytes())
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
