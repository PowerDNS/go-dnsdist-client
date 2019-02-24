package dnsdist

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/crypto/nacl/secretbox"
	"io"
	"log"
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
	fmt.Println("key", dc.key)
	encodedcommand := make([]byte, 0)
	encodedcommand = secretbox.Seal(encodedcommand, []byte(cmd), &dc.writingNonce, &dc.key)
	incrementNonce(&dc.writingNonce)

	fmt.Println("encodedcommand", encodedcommand)
	sendlen := make([]byte, 4)
	binary.BigEndian.PutUint32(sendlen, uint32(len(encodedcommand)))
	n3, err := dc.conn.Write(sendlen)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("wrote", n3, "bytes")
	n4, err := dc.conn.Write(encodedcommand)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("wrote", n4, "bytes")

	recvlenbuf := make([]byte, 4)
	n5, err := io.ReadFull(dc.conn, recvlenbuf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("read", n5, "bytes")
	recvlen := binary.BigEndian.Uint32(recvlenbuf)
	fmt.Println("should read", recvlen, "bytes")
	recvbuf := make([]byte, recvlen)
	n6, err := io.ReadFull(dc.conn, recvbuf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("read", n6, "bytes")
	decodedresponse := make([]byte, 0)
	decodedresponse, ok := secretbox.Open(decodedresponse, recvbuf, &dc.readingNonce, &dc.key)
	incrementNonce(&dc.readingNonce)
	if !ok {
		log.Fatal("secretbox")
	}
	fmt.Println("response:", string(decodedresponse))
	return string(decodedresponse), nil
}

func Dial(target string, secret string) (*DnsdistConn, error) {
	ourNonce := make([]byte, 24)
	rand.Read(ourNonce)
	fmt.Println("ourNonce", ourNonce)

	conn, err := net.Dial("tcp", target)
	if err != nil {
		log.Fatal(err)
	}
	n, err := conn.Write(ourNonce)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("wrote", n, "bytes")
	theirNonce := make([]byte, 24)
	n2, err := io.ReadFull(conn, theirNonce)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("got", n2, "bytes")
	fmt.Println("theirNonce", theirNonce)

	if len(ourNonce) != len(theirNonce) {
		log.Fatal("Received a nonce of size", len(theirNonce), ",  expecting ", len(ourNonce))
	}

	var readingNonce [24]byte
	copy(readingNonce[0:12], ourNonce[0:12])
	copy(readingNonce[12:], theirNonce[12:])
	fmt.Println("readingNonce", readingNonce)

	var writingNonce [24]byte
	copy(writingNonce[0:12], theirNonce[0:12])
	copy(writingNonce[12:], ourNonce[12:])
	fmt.Println("writingNonce", writingNonce)

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
