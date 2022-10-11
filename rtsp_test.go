package main

import (
	"fmt"
	"net"
	"testing"
)

func TestEcho(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:6541")
	if err != nil {
		fmt.Errorf("%s", err.Error())
	}

	write, err := conn.Write([]byte("open"))
	if err != nil {
		return
	}
	fmt.Println(write)
}
