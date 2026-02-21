package model

type ExternalAccountInquiryAdditionalInfo struct {
	// BCA
	InquiryService  string
	SourceAccountNo string
	Amount          *Amount
	PurposeCode     string
	// BRI
	ServiceCode string
	DeviceId    string
	Channel     string
}

type ExternalAccountInquiryRequest struct {
	PartnerId  string
	ChannelId  string
	ExternalId string
	Origin     string

	BeneficiaryBankCode  string
	BeneficiaryAccountNo string
	PartnerReferenceNo   string
	AdditionalInfo       *ExternalAccountInquiryAdditionalInfo
}

type ExternalAccountInquiryResponse struct {
	PartnerReferenceNo     string
	ResponseCode           string
	ResponseMessage        string
	ReferenceNo            string
	BeneficiaryAccountName string
	BeneficiaryAccountNo   string
	BeneficiaryBankCode    string
	BeneficiaryBankName    string
	Currency               string
	Raw                    interface{}
}
