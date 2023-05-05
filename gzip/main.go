package main

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func main() {
	//bytes, err := ioutil.ReadFile("./resp.txt")
	//if err != nil {
	//	panic(err)
	//}
	//data, err := ParseGzip(bytes)
	//fmt.Println(data)
	//fmt.Println(err)

	data, _ := GzipEncode([]byte("hello world"))
	marshal, _ := json.Marshal(string(data))
	fmt.Println(string(marshal))
	//fmt.Println(string(data))
}

func ParseGzip(data []byte) ([]byte, error) {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, data)
	r, err := gzip.NewReader(b)
	if err != nil {
		//logger.Info("[ParseGzip] NewReader error: %v, maybe data is ungzip", err)
		return nil, err
	} else {
		defer r.Close()
		undatas, err := ioutil.ReadAll(r)
		if err != nil {
			//logger.Warn("[ParseGzip]  ioutil.ReadAll error: %v", err)
			return nil, err
		}
		return undatas, nil
	}
}

func GzipEncode(in []byte) ([]byte, error) {
	var (
		buffer bytes.Buffer
		out    []byte
		err    error
	)
	writer := gzip.NewWriter(&buffer)
	_, err = writer.Write(in)
	if err != nil {
		writer.Close()
		return out, err
	}
	err = writer.Close()
	if err != nil {
		return out, err
	}

	return buffer.Bytes(), nil
}

func GzipDecode(in []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		var out []byte
		return out, err
	}
	defer reader.Close()

	return ioutil.ReadAll(reader)
}
