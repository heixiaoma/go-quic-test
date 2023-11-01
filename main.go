package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/quic-go/quic-go"
	"io"
	"math/big"
	"os"
	"time"
)

const addr = "localhost:4242"

// Replace your test file
const bigFilePath = "/xxx/xxxx/xxx.mp4"

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"HP_PRO"},
	}
}
func server() {
	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	if err != nil {
		println(err.Error())
	}
	conn, err := listener.Accept(context.Background())
	if err != nil {
		println(err.Error())

	}
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		println(err.Error())
	}
	reader := bufio.NewReader(stream)
	start := time.Now()

	for {
		_, err := reader.Peek(1)
		if err != nil {
			break
		}
		data := make([]byte, reader.Buffered())
		io.ReadFull(reader, data)
	}
	elapsed := time.Since(start)
	fmt.Printf("Time taken to receive：%s\n", elapsed)
}

func client() {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"HP_PRO"},
	}
	conn, err := quic.DialAddr(context.Background(), addr, tlsConf, nil)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Printf(err.Error())
	}

	fmt.Println("Client: Sending File")

	file, err := os.Open(bigFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	buffer := make([]byte, 2048)
	start := time.Now()
	for {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			fmt.Println(err)
			break
		}
		if bytesRead == 0 {
			break
		}
		_, err = stream.Write(buffer)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("It takes time to send：%s\n", elapsed)
	stream.Close()
	if err != nil {
		fmt.Printf(err.Error())
	}

}

func main() {
	go server()
	go client()
	select {}
}
