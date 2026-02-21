package model

type QRMPMInquiryAdditionalInfo struct {
	TerminalId          string
	PartnerMerchantType string
}

type QRMPMInquiryRequest struct {
	OriginalPartnerReferenceNo string
	OriginalReferenceNo        string
	ServiceCode                string
	MerchantId                 string
	SubMerchantId              string
	AdditionalInfo             *QRMPMInquiryAdditionalInfo
	PartnerId                  string
	ChannelId                  string
	ExternalId                 string
}

type QRMPMInquiryMerchantInfo struct {
	MerchantId         string
	MerchantPan        string
	Name               string
	City               string
	PostalCode         string
	Country            string
	Email              *string
	PaymentChannelName string
}

type QRMPMInquiryResponseAdditionalInfo struct {
	ReferenceNumber       string
	ApprovalCode          *string
	PayerPhoneNumber      *string
	BatchNumber           *string
	ConvenienceFee        *string
	IssuerReferenceNumber *string
	PayerName             *string
	CustomerPan           *string
	IssuerName            *string
	AcquirerName          *string
	MerchantInfo          *QRMPMInquiryMerchantInfo
}

type QRMPMInquiryResponse struct {
	ResponseCode               string
	ResponseMessage            string
	OriginalPartnerReferenceNo string
	OriginalReferenceNo        string
	OriginalExternalId         string
	ServiceCode                string
	LatestTransactionStatus    string
	TransactionStatusDesc      string
	PaidTime                   *string
	Amount                     *Amount
	FeeAmount                  *Amount
	TerminalId                 *string
	AdditionalInfo             *QRMPMInquiryResponseAdditionalInfo
	Raw                        interface{}
}
