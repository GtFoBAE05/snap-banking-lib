package model

import "time"

type QRISNotificationRequest struct {
	OriginalReferenceNo        string
	OriginalPartnerReferenceNo string
	LatestTransactionStatus    string
	TransactionStatusDesc      string
	CustomerNumber             string
	AccountType                *string
	DestinationNumber          string
	DestinationAccountName     string
	Amount                     *Amount
	BankCode                   *string
	AdditionalInfo             *QRISNotificationAdditionalInfo
	PartnerId                  string
	ChannelId                  string
	ExternalId                 string
}

type QRISNotificationMerchantInfo struct {
	TerminalId         string
	MerchantId         string
	City               string
	PostalCode         string
	Country            string
	Email              *string
	PaymentChannelName string
}

type QRISNotificationAdditionalInfo struct {
	ReferenceNumber       string
	TransactionDate       time.Time
	ApprovalCode          string
	PayerPhoneNumber      string
	BatchNumber           string
	ConvenienceFee        string
	IssuerReferenceNumber string
	PayerName             string
	IssuerName            string
	AcquirerName          string
	MerchantInfo          *QRISNotificationMerchantInfo
}

type QRISNotificationResponse struct {
	ResponseCode    string
	ResponseMessage string
}
