package agent

import (
	"bytes"
	"fmt"
)

// HouseSeller 房屋出售者
type HouseSeller interface {
	SellHouse(address string, buyer string) string
}

// houseProxy 房产中介代理
type houseProxy struct {
	houseSeller HouseSeller
}

func NewHouseProxy(houseSeller HouseSeller) *houseProxy {
	return &houseProxy{
		houseSeller: houseSeller,
	}
}

// SellHouse 中介卖房，看房->初步谈价->最终和房东签协议
func (h *houseProxy) SellHouse(address string, buyer string) string {
	buf := bytes.Buffer{}
	buf.WriteString(h.viewHouse(address, buyer) + "\n")
	buf.WriteString(h.preBargain(address, buyer) + "\n")
	// 代理之后再调用最终对象
	buf.WriteString(h.houseSeller.SellHouse(address, buyer))
	return buf.String()
}

// viewHouse 看房介绍基本情况
func (h *houseProxy) viewHouse(address string, buyer string) string {
	return fmt.Sprintf("带买家%s看位于%s的房屋，并介绍基本情况", buyer, address)
}

// preBargain 初步沟通价格
func (h *houseProxy) preBargain(address string, buyer string) string {
	return fmt.Sprintf("讨价还价后，初步达成购买意向")
}

// houseOwner 房东
type houseOwner struct{}

// SellHouse 房东卖房，商讨价格，签署购房协议
func (h *houseOwner) SellHouse(address string, buyer string) string {
	return fmt.Sprintf("最终商讨价格后，与%s签署购买地址为%s的购房协议。", buyer, address)
}
