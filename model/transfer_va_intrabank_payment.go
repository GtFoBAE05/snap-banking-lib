package model

import "time"

type VirtualAccountIntrabankPaymentRequest struct {
	PartnerId  string
	ChannelId  string
	ExternalId string

	PartnerServiceId    string
	CustomerNo          string
	VirtualAccountNo    string
	VirtualAccountName  string
	SourceAccountNo     string
	PartnerReferenceNo  string
	PaidAmount          *Amount
	TrxDateTime         time.Time
	VirtualAccountEmail string
}

type VirtualAccountIntrabankPaymentResponse struct {
	ResponseCode       string
	ResponseMessage    string
	VirtualAccountData *VirtualAccountIntrabankPaymentData
	Raw                interface{}
}

type VirtualAccountIntrabankPaymentData struct {
	PartnerServiceId    string
	CustomerNo          string
	VirtualAccountNo    string
	VirtualAccountName  string
	PartnerReferenceNo  string
	PaymentRequestId    string
	PaidAmount          *Amount
	TrxDateTime         string
	VirtualAccountEmail string
	SourceAccountNo     string
	ReferenceNo         string
	TotalAmount         *Amount
	BillDetails         []IntrabankBillDetail
	FreeTexts           []LocalizedText
	FeeAmount           *Amount
	ProductName         string
}
