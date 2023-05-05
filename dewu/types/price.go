package types

type Price struct {
	Code   int    `json:"code"`
	Status bool   `json:"status"`
	Msg    string `json:"msg"`
	Data   struct {
		Msg  string `json:"msg"`
		Code int    `json:"code"`
		Data struct {
			PriceDescriptionFlag    bool   `json:"priceDescriptionFlag"`
			IntensifyNinetyFiveFlag bool   `json:"intensifyNinetyFiveFlag"`
			ShowOpenBtnText         string `json:"showOpenBtnText"`
			TimeThreshold           int    `json:"timeThreshold"`
			InstallmentExtend       struct {
				OldCustomer bool `json:"oldCustomer"`
			} `json:"installmentExtend"`
			TotalMinPrice    int   `json:"totalMinPrice"`
			DayTimeThreshold int   `json:"dayTimeThreshold"`
			SpuHasBids       bool  `json:"spuHasBids"`
			SpuId            int   `json:"spuId"`
			ServerTime       int64 `json:"serverTime"`
			SkuInfoList      []struct {
				OptimalDiscountInfo struct {
					ActivePrice   int   `json:"activePrice"`
					OriginalPrice int   `json:"originalPrice"`
					TradeTypes    []int `json:"tradeTypes"`
					Price         int   `json:"price"`
				} `json:"optimalDiscountInfo,omitempty"`
				MinPrice             int `json:"minPrice,omitempty"`
				TradeChannelInfoList []struct {
					ArrivalTimeText        string `json:"arrivalTimeText"`
					OptimalPrice           int    `json:"optimalPrice,omitempty"`
					SaleInventoryNo        string `json:"saleInventoryNo,omitempty"`
					Price                  int    `json:"price"`
					ChannelName            string `json:"channelName,omitempty"`
					TradeDesc              string `json:"tradeDesc"`
					ItemDetailAgingMaxTime int    `json:"itemDetailAgingMaxTime,omitempty"`
					TradeType              int    `json:"tradeType"`
					BidType                int    `json:"bidType,omitempty"`
					AgingExtInfo           struct {
						DataSource int      `json:"dataSource"`
						ParkCodes  []string `json:"parkCodes,omitempty"`
					} `json:"agingExtInfo,omitempty"`
					GlobalText             string `json:"globalText,omitempty"`
					ChannelAdditionInfoDTO struct {
						Explain string `json:"explain"`
						Symbol  string `json:"symbol"`
					} `json:"channelAdditionInfoDTO,omitempty"`
					FlagUrl     string `json:"flagUrl,omitempty"`
					Background  string `json:"background,omitempty"`
					LinkUrl     string `json:"linkUrl,omitempty"`
					ActivePrice int    `json:"activePrice,omitempty"`
				} `json:"tradeChannelInfoList"`
				SkuAdditionInfoDTO struct {
					DiscountText string `json:"discountText,omitempty"`
					PreTaxText   string `json:"preTaxText,omitempty"`
					AfterTaxText string `json:"afterTaxText,omitempty"`
				} `json:"skuAdditionInfoDTO,omitempty"`
				PostageDiscountDTO struct {
					DiscountAmt     int      `json:"discountAmt"`
					ExpireTime      int64    `json:"expireTime"`
					NewUserLabel    bool     `json:"newUserLabel"`
					StartTime       int64    `json:"startTime"`
					Label           string   `json:"label"`
					Title           string   `json:"title"`
					Type            int      `json:"type"`
					RuleDescription []string `json:"ruleDescription"`
				} `json:"postageDiscountDTO,omitempty"`
				SkuDiscountLabel struct {
					LabelName string `json:"labelName"`
				} `json:"skuDiscountLabel,omitempty"`
				TradeType       int `json:"tradeType,omitempty"`
				SkuId           int `json:"skuId"`
				InstallmentInfo struct {
					TrialText       string `json:"trialText"`
					InstalmentText  string `json:"instalmentText"`
					FreeText        string `json:"freeText"`
					SptHbfq         bool   `json:"sptHbfq"`
					TrialLayerText  string `json:"trialLayerText"`
					InstallmentType int    `json:"installmentType"`
					SptJwfq         bool   `json:"sptJwfq"`
				} `json:"installmentInfo,omitempty"`
				CanOpenDiscountFloat bool `json:"canOpenDiscountFloat,omitempty"`
			} `json:"skuInfoList"`
			CanShowArrivalBtn bool `json:"canShowArrivalBtn"`
		} `json:"data"`
		Status int `json:"status"`
	} `json:"data"`
	Ts int64 `json:"ts"`
}
