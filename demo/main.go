package main

type A struct {
	Name string
}

func (i *A) SetName(name string)  {
	i.Name = name
}

func (i *A) GetName()  {

}


func main() {
	name := "123"
	var a interface{}
	a = A{}
	if aa, ok := a.(interface{SetName(name string)}); ok {
		aa.SetName(name)
	}
}

