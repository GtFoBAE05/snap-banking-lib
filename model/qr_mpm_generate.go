package model

import "time"

type QRMPMGenerateAdditionalInfo struct {
	ConvenienceFee       string
	PartnerMerchantType  string
	TerminalLocationName string
	QrOption             string
}

type QRMPMGenerateRequest struct {
	PartnerReferenceNo string
	Amount             *Amount
	MerchantId         string
	SubMerchantId      string
	TerminalId         string
	ValidityPeriod     time.Time
	AdditionalInfo     *QRMPMGenerateAdditionalInfo
	PartnerId          string
	ChannelId          string
	ExternalId         string
}

type QRMPMGenerateResponse struct {
	ResponseCode       string
	ResponseMessage    string
	ReferenceNo        string
	PartnerReferenceNo string
	QrContent          string
	QrUrl              *string
	QrImage            string
	RedirectUrl        *string
	MerchantName       string
	StoreId            *string
	TerminalId         string
	AdditionalInfo     interface{}
	Raw                interface{}
}
