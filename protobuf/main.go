package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
)

// 序列化以后数据被转换成二进制类型，反序列化后直接写入对应的结构体类型。
func main()  {
	nxin := Nxin{
		Id: *proto.Int32(1),
		Name: *proto.String("joy"),
		Telphone: *proto.String("13000000000"),
	}
	bytes, err := proto.Marshal(&nxin) // 序列化
	fmt.Println(string(bytes))
	fmt.Println(err)
	n := Nxin{}
	err = proto.Unmarshal(bytes, &n) // 反序列化
	fmt.Println(n)
	fmt.Println(err)
	fmt.Println(n.GetName())
	//f,_:=os.Create("./dx.txt")
	//defer f.Close()
	//f.Write(bytes)
}
