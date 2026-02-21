package model

type BalanceInquiryRequest struct {
	PartnerId  string
	ChannelId  string
	ExternalId string

	AccountNo          string
	BankCardToken      string
	PartnerReferenceNo string
}

type BalanceInquiryResponse struct {
	ResponseCode             string
	ResponseMessage          string
	ReferenceNo              string
	PartnerReferenceNo       string
	AccountNo                string
	Name                     string
	Amount                   *Amount
	FloatAmount              *Amount
	HoldAmount               *Amount
	AvailableBalance         *Amount
	LedgerBalance            *Amount
	CurrentMultilateralLimit *Amount
	Status                   string
	ProductCode              string
	ProductDesc              string
	AccountType              string
	Raw                      interface{}
}
