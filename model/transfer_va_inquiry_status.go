package model

type VirtualAccountInquiryStatusRequest struct {
	PartnerId  string
	ChannelId  string
	ExternalId string

	PartnerServiceId string
	CustomerNo       string
	VirtualAccountNo string
	PaymentRequestId string
	AdditionalInfo   map[string]interface{}
}

type VirtualAccountInquiryStatusResponse struct {
	ResponseCode       string
	ResponseMessage    string
	VirtualAccountData *VirtualAccountInquiryStatusData
	Raw                interface{}
}

type VirtualAccountInquiryStatusData struct {
	PaymentFlagStatus string
	PaymentFlagReason *LocalizedText
	PartnerServiceId  string
	CustomerNo        string
	VirtualAccountNo  string
	InquiryRequestId  string
	PaymentRequestId  string
	PaidAmount        *Amount
	TotalAmount       *Amount
	TransactionDate   string
	ReferenceNo       string
	BillDetails       []BillDetail
}
