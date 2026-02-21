package model

import "time"

type VirtualAccountInquiryRequest struct {
	PartnerId  string
	ChannelId  string
	ExternalId string

	PartnerServiceId string
	CustomerNo       string
	VirtualAccountNo string
	TrxDateInit      time.Time
	ChannelCode      int
	AdditionalInfo   map[string]interface{}
	InquiryRequestId string
}

type VirtualAccountInquiryResponse struct {
	ResponseCode       string
	ResponseMessage    string
	VirtualAccountData *VirtualAccountData
	Raw                interface{}
}

type VirtualAccountData struct {
	InquiryStatus         string
	InquiryReason         *LocalizedText
	PartnerServiceId      string
	CustomerNo            string
	VirtualAccountNo      string
	VirtualAccountName    string
	InquiryRequestId      string
	TotalAmount           *Amount
	SubCompany            string
	BillDetails           []BillDetail
	FreeTexts             []LocalizedText
	VirtualAccountTrxType string
	FeeAmount             *Amount
	AdditionalInfo        map[string]AdditionalInfoItem
}
