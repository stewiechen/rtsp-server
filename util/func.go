package util

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"regexp"
	"time"
)

func RegTo(buf, reg string) []string {
	var res []string

	re, _ := regexp.Compile(reg)
	restmp := re.FindAllStringSubmatch(buf, -1)

	for _, tmp := range restmp {
		for _, val := range tmp {
			res = append(res, val)
		}
	}

	return res
}

func BytesToInt8(data []byte) int {
	binbuf := bytes.NewBuffer(data)
	var x int8
	_ = binary.Read(binbuf, binary.BigEndian, &x)
	return int(x)
}

func BytesToInt16(bys []byte) int {
	binbuf := bytes.NewBuffer(bys)
	var x int16
	_ = binary.Read(binbuf, binary.BigEndian, &x)
	return int(x)
}

func RandomString(l int) string {
	bytes0 := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	var result []byte

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes0[r.Intn(len(bytes0))])
	}

	return string(result)
}

func BytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}
