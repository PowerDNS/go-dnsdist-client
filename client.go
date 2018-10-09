package main

import (
	"encoding/base64"
	"io"
	"fmt"
	"encoding/binary"
	"log"
	"net"
	"github.com/jamesruan/sodium"
    // "golang.org/x/crypto/nacl/secretbox"
)

func main() {
	ourNonce := sodium.SecretBoxNonce{}
	sodium.Randomize(&ourNonce)
	fmt.Println("ourNonce", ourNonce)

	conn, err := net.Dial("tcp", "127.0.0.1:5199")
	if err != nil {
		log.Fatal(err)
	}
	n, err := conn.Write(ourNonce.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("wrote", n, "bytes")
	// theirNonce := make([]byte, 0, len(ourNonce.Bytes))
	theirNonce := sodium.SecretBoxNonce{}
	theirNonce.Bytes = make([]byte, len(ourNonce.Bytes))
	fmt.Println("len(ourNonce.Bytes)=", len(ourNonce.Bytes))
	fmt.Println("len(theirNonce.Bytes)=", len(theirNonce.Bytes))
	n2, err := io.ReadFull(conn, theirNonce.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("got", n2, "bytes")
	fmt.Println("theirNonce", theirNonce)

	if len(ourNonce.Bytes) != len(theirNonce.Bytes) {
		log.Fatal("Received a nonce of size", len(theirNonce.Bytes),",  expecting ", len(ourNonce.Bytes))
	}

	halfNonceSize := int(len(ourNonce.Bytes)/2)
	fmt.Println("halfNonceSize=", halfNonceSize)
	readingNonce := sodium.SecretBoxNonce{}
	writingNonce := sodium.SecretBoxNonce{}

	readingNonce.Bytes = make([]byte, halfNonceSize)
	copy(readingNonce.Bytes, ourNonce.Bytes[0:halfNonceSize])
	readingNonce.Bytes = append(readingNonce.Bytes, theirNonce.Bytes[halfNonceSize:]...)
	fmt.Println("readingNonce", readingNonce)

	writingNonce.Bytes = make([]byte, halfNonceSize)
	copy(writingNonce.Bytes, theirNonce.Bytes[0:halfNonceSize])
	writingNonce.Bytes = append(writingNonce.Bytes, ourNonce.Bytes[halfNonceSize:]...)
	fmt.Println("writingNonce", writingNonce)

	fmt.Println("ourNonce", ourNonce)
	fmt.Println("theirNonce", theirNonce)

	command := sodium.Bytes([]byte("print(123)"))
	key := sodium.SecretBoxKey{}
	key.Bytes, err = base64.StdEncoding.DecodeString("WQcBTlKzEuTbMTdydMSW1CSQvyIAINML6oIGfGOjXjE=")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("key", key)
	encodedcommand := command.SecretBox(writingNonce, key)

	fmt.Println("encodedcommand", encodedcommand)
	sendlen := make([]byte, 4)
	binary.BigEndian.PutUint32(sendlen, uint32(len(encodedcommand)))
	n3, err := conn.Write(sendlen)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("wrote", n3, "bytes")
	n4, err := conn.Write(encodedcommand)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("wrote", n4, "bytes")

	// var buf []byte
	// var nonce [24]byte
	// nonce = writingNonce.Bytes[:]
	// encrypted := secretbox.Seal(buf, command, &nonce, &key.Bytes)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("nacl sealed", encrypted)
}