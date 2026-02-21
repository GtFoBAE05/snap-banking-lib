package model

type InternalAccountInquiryRequest struct {
	PartnerId  string
	ChannelId  string
	ExternalId string
	Origin     string

	PartnerReferenceNo   string
	BeneficiaryAccountNo string
}

type InternalAccountInquiryResponse struct {
	ResponseCode           string
	ResponseMessage        string
	ReferenceNo            string
	PartnerReferenceNo     string
	BeneficiaryAccountName string
	BeneficiaryAccountNo   string

	// BRI
	BeneficiaryAccountStatus string
	BeneficiaryAccountType   string
	Currency                 string
	Raw                      interface{}
}
