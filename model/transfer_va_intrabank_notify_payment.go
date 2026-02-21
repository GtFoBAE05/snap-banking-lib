package model

import "time"

type VirtualAccountIntrabankPaymentNotificationAdditionalInfo struct {
	IdApp         string
	PassApp       string
	PaymentAmount string
	TerminalId    string
	BankId        string
}

type VirtualAccountIntrabankPaymentNotificationRequest struct {
	PartnerId  string
	ChannelId  string
	ExternalId string

	PartnerServiceId   string
	CustomerNo         string
	VirtualAccountNo   string
	PaymentRequestId   string
	TrxDateTime        time.Time
	AdditionalInfo     *VirtualAccountIntrabankPaymentNotificationAdditionalInfo
	PartnerReferenceNo string
	PaymentStatus      string
	PaymentFlagReason  *LocalizedText
}

type VirtualAccountIntrabankPaymentNotificationResponse struct {
	ResponseCode       string
	ResponseMessage    string
	VirtualAccountData *VirtualAccountIntrabankPaymentNotificationData
	Raw                interface{}
}

type VirtualAccountIntrabankPaymentNotificationData struct {
	PartnerServiceId   string
	CustomerNo         string
	VirtualAccountNo   string
	InquiryRequestId   string
	PaymentRequestId   string
	TrxDateTime        string
	PaymentStatus      string
	PartnerReferenceNo string
}
