package model

import "time"

type InterbankTransferAdditionalInfo struct {
	// BCA
	TransferType string
	PurposeCode  string

	// BRI
	ServiceCode          string
	DeviceId             string
	Channel              string
	ReferenceNo          string
	ExternalId           string
	SenderIdentityNumber string
	PaymentInfo          string
	SenderType           string
	SenderResidentStatus string
	SenderTownName       string
	IsRdn                string
}

type InterbankTransferRequest struct {
	PartnerId  string
	ChannelId  string
	ExternalId string

	PartnerReferenceNo     string
	Amount                 *Amount
	BeneficiaryAccountName string
	BeneficiaryAccountNo   string
	BeneficiaryAddress     string
	BeneficiaryBankCode    string
	BeneficiaryBankName    string
	BeneficiaryEmail       string
	CustomerReference      string
	SourceAccountNo        string
	TransactionDate        time.Time
	AdditionalInfo         *InterbankTransferAdditionalInfo
	OriginatorInfos        []OriginatorInfo
}

type InterbankTransferResponse struct {
	ResponseCode         string
	ResponseMessage      string
	ReferenceNo          string
	PartnerReferenceNo   string
	Amount               *Amount
	BeneficiaryAccountNo string
	BeneficiaryBankCode  string
	SourceAccountNo      string
	OriginatorInfos      []OriginatorInfo
	Raw                  interface{}
}
