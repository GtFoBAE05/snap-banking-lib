package model

import "time"

type VirtualAccountPaymentRequest struct {
	PartnerId  string
	ChannelId  string
	ExternalId string

	PartnerServiceId        string
	CustomerNo              string
	VirtualAccountNo        string
	VirtualAccountName      string
	PaymentRequestId        string
	ChannelCode             int
	HashedSourceAccountNo   string
	SourceBankCode          string
	PaidAmount              *Amount
	CumulativePaymentAmount *Amount
	PaidBills               string
	TotalAmount             *Amount
	TrxDateTime             time.Time
	ReferenceNo             string
	FlagAdvise              string
	SubCompany              string
	BillDetails             []BillDetail
	AdditionalInfo          map[string]interface{}
}

type VirtualAccountPaymentResponse struct {
	ResponseCode       string
	ResponseMessage    string
	VirtualAccountData *VirtualAccountPaymentData
	AdditionalInfo     map[string]AdditionalInfoItem
	Raw                interface{}
}

type VirtualAccountPaymentData struct {
	PaymentFlagReason  *LocalizedText
	PartnerServiceId   string
	CustomerNo         string
	VirtualAccountNo   string
	VirtualAccountName string
	PaymentRequestId   string
	PaidAmount         *Amount
	TotalAmount        *Amount
	TrxDateTime        string
	ReferenceNo        string
	PaymentFlagStatus  string
	BillDetails        []BillDetail
}
