package main

import (
	"encoding/json"
	"fmt"
	"go_master/dewu/types"
	"os"
	"strings"
)

type Sku struct {
	SpuId         int          `json:"spuId"`
	SkuId         int          `json:"skuId"`
	AuthPrice     int          `json:"authPrice"`
	ArticleNumber string       `json:"articleNumber"`
	Title         string       `json:"title"`
	Properties    []properties `json:"properties"` // 销售属性数组 如尺寸 颜色等
}

type properties struct {
	PropertyValueId int    `json:"propertyValueId"`
	Value           string `json:"value"`
	Name            string `json:"name"`
	Price           int    `json:"price"`
	OriginalPrice   int    `json:"originalPrice"`
	ActivePrice     int    `json:"activePrice"`
}

type PriceSku struct {
	MinPrice      int `json:"minPrice"`
	OriginalPrice int `json:"originalPrice"`
	ActivePrice   int `json:"activePrice"`
}

func main() {
	detail := loadDetail()
	price := loadPrice()
	// 概念 spu 是一个商品的概念，比如一双鞋子，一件衣服
	// 概念 sku 是一个商品的具体的某个规格，比如一双鞋子的 42 码，一件衣服的红色
	// 货号 sku id 尺码 价格
	// price 中 skuId => price
	// detail 中 skuId => propertyValueId => name 尺码 的 value
	skus := detail.Data.Data.Skus
	// finish response map
	data := make(map[int]Sku)
	for _, sku := range skus {
		i := []properties{}
		for _, prop := range sku.Properties {
			i = append(i, properties{
				PropertyValueId: prop.PropertyValueId,
				Value:           "0",
				Name:            "待补充",
				Price:           0,
			})
		}
		data[sku.SkuId] = Sku{
			SpuId:         sku.SpuId,
			SkuId:         sku.SkuId,
			AuthPrice:     sku.AuthPrice,
			Properties:    i,
			ArticleNumber: detail.Data.Data.Detail.ArticleNumber,
			Title:         detail.Data.Data.Detail.Title,
		}
	}
	// saleProperties map
	propertiesMap := make(map[int]properties)
	for _, prop := range detail.Data.Data.SaleProperties.List {
		propertiesMap[prop.PropertyValueId] = properties{
			PropertyValueId: prop.PropertyValueId,
			Value:           prop.Value,
			Name:            prop.Name,
		}
	}
	// price map
	priceMap := make(map[int]PriceSku)
	for _, sku := range price.Data.Data.SkuInfoList {
		priceMap[sku.SkuId] = PriceSku{
			MinPrice:      sku.MinPrice,
			OriginalPrice: sku.OptimalDiscountInfo.OriginalPrice,
			ActivePrice:   sku.OptimalDiscountInfo.ActivePrice,
		}
	}
	// 补充数据
	for k, sku := range data {
		// 补充spu的name 和 value
		for pk, prop := range sku.Properties {
			if v, ok := propertiesMap[prop.PropertyValueId]; ok {
				data[k].Properties[pk].Value = v.Value
				data[k].Properties[pk].Name = v.Name
			}
			if v, ok := priceMap[k]; ok {
				data[k].Properties[pk].Price = v.MinPrice
				data[k].Properties[pk].OriginalPrice = v.OriginalPrice
				data[k].Properties[pk].ActivePrice = v.ActivePrice
			}
		}
	}
	// 重组成数组
	res := make([]Sku, 0)
	for _, d := range data {
		res = append(res, d)
	}
	bytes, _ := json.Marshal(res)
	s := string(bytes)
	fmt.Println(s)
}

func loadDetail() *types.Detail {
	filePath := "./_data/detail.json"
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	raw := string(bytes)
	raw = strings.Replace(raw, "\uFEFF", "", -1)

	var detail types.Detail
	err = json.Unmarshal([]byte(raw), &detail)
	if err != nil {
		panic(err)
	}
	return &detail
}

func loadPrice() *types.Price {
	filePath := "./_data/price.json"
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	raw := string(bytes)
	raw = strings.Replace(raw, "\uFEFF", "", -1)
	var price types.Price
	err = json.Unmarshal([]byte(raw), &price)
	if err != nil {
		panic(err)
	}
	return &price
}
