package model

import "time"

type TransactionStatusInquiryRequest struct {
	PartnerId  string
	ChannelId  string
	ExternalId string

	OriginalPartnerReferenceNo string
	OriginalExternalId         string
	ServiceCode                string
	TransactionDate            time.Time
}

type TransactionStatusInquiryResponse struct {
	ResponseCode               string
	ResponseMessage            string
	OriginalReferenceNo        string
	OriginalPartnerReferenceNo string
	OriginalExternalId         string
	ServiceCode                string
	TransactionDate            string
	Amount                     *Amount
	BeneficiaryAccountNo       string
	BeneficiaryBankCode        string
	ReferenceNumber            string
	SourceAccountNo            string
	LatestTransactionStatus    string
	TransactionStatusDesc      string
	Raw                        interface{}
}
