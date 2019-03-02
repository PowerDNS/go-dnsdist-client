package dnsdist

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/nacl/secretbox"
)

type DnsdistConn struct {
	conn         net.Conn
	readingNonce [24]byte
	writingNonce [24]byte
	key          [32]byte
}

func Dial(target string, secret string) (*DnsdistConn, error) {
	ourNonce := make([]byte, 24)
	_, err := rand.Read(ourNonce)
	if err != nil {
		return nil, fmt.Errorf("during dnsdist.Dial: %s", err)
	}

	var key [32]byte
	xkey, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, fmt.Errorf("while decoding shared secret: %s", err)
	}
	if len(xkey) != 32 {
		return nil, fmt.Errorf("shared secret is %v bytes, should be 32", len(xkey))
	}
	copy(key[0:32], xkey)

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
		return nil, fmt.Errorf("received a nonce of size %v, expecting %v", len(theirNonce), len(ourNonce))
	}

	var readingNonce [24]byte
	copy(readingNonce[:12], ourNonce[:12])
	copy(readingNonce[12:], theirNonce[12:])

	var writingNonce [24]byte
	copy(writingNonce[:12], theirNonce[:12])
	copy(writingNonce[12:], ourNonce[12:])

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

func incrementNonce(nonce *[24]byte) {
	value := binary.BigEndian.Uint32(nonce[:4])
	value = value + 1
	binary.BigEndian.PutUint32(nonce[:4], value)
}

func (dc *DnsdistConn) Command(cmd string) (string, error) {
	encodedcommand := secretbox.Seal(nil, []byte(cmd), &dc.writingNonce, &dc.key)
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
	decodedresponse, ok := secretbox.Open(nil, recvbuf, &dc.readingNonce, &dc.key)
	incrementNonce(&dc.readingNonce)
	if !ok {
		return "", fmt.Errorf("error decoding reply")
	}
	return string(decodedresponse), nil
}
