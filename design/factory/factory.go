package factory

// Pancake 煎饼
type Pancake interface {
	// ShowFlour 煎饼使用的面粉
	ShowFlour() string
	// Value 煎饼价格
	Value() float32
}

// PancakeCook 煎饼厨师
type PancakeCook interface {
	// MakePancake 摊煎饼
	MakePancake() Pancake
}

// PancakeVendor 煎饼小贩
type PancakeVendor struct {
	PancakeCook
}

// NewPancakeVendor ...
func NewPancakeVendor(cook PancakeCook) *PancakeVendor {
	return &PancakeVendor{
		PancakeCook: cook,
	}
}

// SellPancake 卖煎饼，先摊煎饼，再卖
func (vendor *PancakeVendor) SellPancake() (money float32) {
	return vendor.MakePancake().Value()
}

// cornPancake 玉米面煎饼
type cornPancake struct{}

// NewCornPancake ...
func NewCornPancake() *cornPancake {
	return &cornPancake{}
}

func (cake *cornPancake) ShowFlour() string {
	return "玉米面"
}

func (cake *cornPancake) Value() float32 {
	return 5.0
}

// milletPancake 小米面煎饼
type milletPancake struct{}

func NewMilletPancake() *milletPancake {
	return &milletPancake{}
}

func (cake *milletPancake) ShowFlour() string {
	return "小米面"
}

func (cake *milletPancake) Value() float32 {
	return 8.0
}

// cornPancakeCook 制作玉米面煎饼厨师
type cornPancakeCook struct{}

func NewCornPancakeCook() *cornPancakeCook {
	return &cornPancakeCook{}
}

func (cook *cornPancakeCook) MakePancake() Pancake {
	return NewCornPancake()
}

// milletPancakeCook 制作小米面煎饼厨师
type milletPancakeCook struct{}

func NewMilletPancakeCook() *milletPancakeCook {
	return &milletPancakeCook{}
}

func (cook *milletPancakeCook) MakePancake() Pancake {
	return NewMilletPancake()
}
