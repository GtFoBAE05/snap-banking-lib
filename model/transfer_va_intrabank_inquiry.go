package model

type VirtualAccountIntrabankInquiryRequest struct {
	PartnerId  string
	ChannelId  string
	ExternalId string

	// BRI
	PartnerServiceId string
	CustomerNo       string
	VirtualAccountNo string

	// BCA
	Amount *Amount
}

type VirtualAccountIntrabankInquiryResponse struct {
	ResponseCode       string
	ResponseMessage    string
	VirtualAccountData *VirtualAccountIntrabankInquiryData
	Raw                interface{}
}

type VirtualAccountIntrabankInquiryData struct {
	// BRI
	PartnerServiceId   string
	CustomerNo         string
	VirtualAccountNo   string
	VirtualAccountName string
	TotalAmount        *Amount

	BillDetails           []IntrabankBillDetail
	FreeTexts             []LocalizedText
	VirtualAccountTrxType string
	FeeAmount             *Amount
	ProductName           string
}

type IntrabankBillDetail struct {
	BillDescription *LocalizedText
	BillAmount      *Amount
}
