package model

import "time"

type BankStatementRequest struct {
	PartnerId  string
	ChannelId  string
	ExternalId string

	AccountNo          string
	PartnerReferenceNo string
	BankCardToken      string
	FromDateTime       time.Time
	ToDateTime         time.Time
}

type BankStatementResponse struct {
	ResponseCode       string
	ResponseMessage    string
	ReferenceNo        string
	PartnerReferenceNo string
	Balance            []Balance
	TotalCreditEntries *TotalEntries
	TotalDebitEntries  *TotalEntries
	DetailData         []TransactionDetail
	Raw                interface{}
}

type Balance struct {
	Amount          *AmountWithDateTime
	StartingBalance *AmountWithDateTime
	EndingBalance   *AmountWithDateTime
}

type AmountWithDateTime struct {
	Value    string
	Currency string
	DateTime time.Time
}

type TotalEntries struct {
	NumberOfEntries string
	Amount          *Amount
}

type TransactionDetail struct {
	Amount          *Amount
	TransactionDate time.Time
	Remark          string
	Type            string

	// BRI-specific
	TransactionId string
	Balance       *Balance
	RemarkCustom  string
}
