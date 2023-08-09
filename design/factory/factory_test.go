package factory

import (
	"fmt"
	"testing"
)

func TestFactoryMethod(t *testing.T) {
	pancakeVendor := NewPancakeVendor(NewCornPancakeCook())
	fmt.Printf("Corn pancake value is %v\n", pancakeVendor.SellPancake())

	pancakeVendor = NewPancakeVendor(NewMilletPancakeCook())
	fmt.Printf("Millet pancake value is %v\n", pancakeVendor.SellPancake())
}
