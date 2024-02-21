package desing

type counter interface {
	Add()
	Sub()
	Multiply()
}

type counterImpl1 struct {
	n int
}

func (c *counterImpl1) Add() {
	c.n++
}

func (c *counterImpl1) Sub() {
	c.n--
}

func (c *counterImpl1) Multiply() {
	c.n *= c.n
}

func addSubMul(c counter) {
	c.Add()
	c.Sub()
	c.Multiply()
}

func addSubMulGenerics[T counter](p T) {
	p.Add()
	p.Sub()
	p.Multiply()
}
