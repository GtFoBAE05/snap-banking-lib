package model

import "time"

type QRMPMRefundAdditionalInfo struct {
	TerminalId          string
	TransactionDate     time.Time
	PartnerMerchantType string
	IssuerName          string
}

type QRMPMRefundRequest struct {
	MerchantId                 string
	OriginalPartnerReferenceNo string
	OriginalReferenceNo        string
	PartnerRefundNo            string
	RefundAmount               *Amount
	AdditionalInfo             *QRMPMRefundAdditionalInfo
	PartnerId                  string
	ChannelId                  string
	ExternalId                 string
}

type QRMPMRefundResponseAdditionalInfo struct {
	MerchantId      string
	TerminalId      string
	ReferenceNumber string
	AvailableAmount *Amount
	RefundCounter   string
}

type QRMPMRefundResponse struct {
	ResponseCode               string
	ResponseMessage            string
	OriginalPartnerReferenceNo string
	OriginalReferenceNo        string
	OriginalExternalId         string
	RefundNo                   string
	PartnerRefundNo            string
	RefundAmount               *Amount
	RefundTime                 string
	AdditionalInfo             *QRMPMRefundResponseAdditionalInfo
	Raw                        interface{}
}
