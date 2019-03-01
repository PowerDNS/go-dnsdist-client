package dnsdist

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/crypto/nacl/secretbox"
	"io"
	"net"
)

type DnsdistConn struct {
	conn         net.Conn
	readingNonce [24]byte
	writingNonce [24]byte
	key          [32]byte
}

func incrementNonce(nonce *[24]byte) {
	value := binary.BigEndian.Uint32(nonce[:4])
	value = value + 1
	binary.BigEndian.PutUint32(nonce[:4], value)
}

func (dc *DnsdistConn) Command(cmd string) (string, error) {
	encodedcommand := make([]byte, 0)
	encodedcommand = secretbox.Seal(encodedcommand, []byte(cmd), &dc.writingNonce, &dc.key)
	incrementNonce(&dc.writingNonce)

	sendlen := make([]byte, 4)
	binary.BigEndian.PutUint32(sendlen, uint32(len(encodedcommand)))
	_, err := dc.conn.Write(sendlen)
	if err != nil {
		return "", fmt.Errorf("while writing encoded command length: %s", err)
	}

	_, err = dc.conn.Write(encodedcommand)
	if err != nil {
		return "", fmt.Errorf("while writing encoded command: %s", err)
	}

	recvlenbuf := make([]byte, 4)
	_, err = io.ReadFull(dc.conn, recvlenbuf)
	if err != nil {
		return "", fmt.Errorf("while reading encoded reply length: %s", err)
	}

	recvlen := binary.BigEndian.Uint32(recvlenbuf)
	recvbuf := make([]byte, recvlen)
	_, err = io.ReadFull(dc.conn, recvbuf)
	if err != nil {
		return "", fmt.Errorf("while reading encoded reply: %s", err)
	}
	decodedresponse := make([]byte, 0)
	decodedresponse, ok := secretbox.Open(decodedresponse, recvbuf, &dc.readingNonce, &dc.key)
	incrementNonce(&dc.readingNonce)
	if !ok {
		return "", fmt.Errorf("error decoding reply")
	}
	return string(decodedresponse), nil
}

func Dial(target string, secret string) (*DnsdistConn, error) {
	ourNonce := make([]byte, 24)
	_, err := rand.Read(ourNonce)
	if err != nil {
		return nil, fmt.Errorf("during dnsdist.Dial: %s", err)
	}

	conn, err := net.Dial("tcp", target)
	if err != nil {
		return nil, fmt.Errorf("during dnsdist.Dial: %s", err)
	}
	_, err = conn.Write(ourNonce)
	if err != nil {
		return nil, fmt.Errorf("error writing our nonce: %s", err)
	}
	theirNonce := make([]byte, 24)
	_, err = io.ReadFull(conn, theirNonce)
	if err != nil {
		return nil, fmt.Errorf("error reading server nonce: %s", err)
	}

	if len(ourNonce) != len(theirNonce) {
		return nil, fmt.Errorf("Received a nonce of size %s, expecting %s", len(theirNonce), len(ourNonce))
	}

	var readingNonce [24]byte
	copy(readingNonce[0:12], ourNonce[0:12])
	copy(readingNonce[12:], theirNonce[12:])

	var writingNonce [24]byte
	copy(writingNonce[0:12], theirNonce[0:12])
	copy(writingNonce[12:], ourNonce[12:])

	var key [32]byte
	xkey, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, fmt.Errorf("while decoding shared secret: %s", err)
	}
	copy(key[0:32], xkey)
	dc := DnsdistConn{conn, readingNonce, writingNonce, key}

	resp, err := dc.Command("")
	if err != nil {
		return nil, err
	}

	if resp != "" {
		return nil, errors.New("handshake error")
	}

	return &dc, nil
}
