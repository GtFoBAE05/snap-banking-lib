package model

import "time"

type IntrabankTransferAdditionalInfo struct {
	// BCA
	EconomicActivity   string
	TransactionPurpose string

	// BRI
	DeviceId string
	Channel  string
	IsRdn    string
}

type IntrabankTransferRequest struct {
	PartnerId  string
	ChannelId  string
	ExternalId string

	PartnerReferenceNo   string
	Amount               *Amount
	BeneficiaryAccountNo string
	BeneficiaryEmail     string
	CustomerReference    string
	FeeType              string
	Remark               string
	SourceAccountNo      string
	TransactionDate      time.Time
	AdditionalInfo       *IntrabankTransferAdditionalInfo
	OriginatorInfos      []OriginatorInfo
}

type IntrabankTransferResponse struct {
	ResponseCode         string
	ResponseMessage      string
	ReferenceNo          string
	PartnerReferenceNo   string
	Amount               *Amount
	BeneficiaryAccountNo string
	CustomerReference    string
	SourceAccountNo      string
	TransactionDate      string
	AdditionalInfo       *IntrabankTransferAdditionalInfo
	OriginatorInfos      []OriginatorInfo
	Raw                  interface{}
}
