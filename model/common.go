package model

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type LocalizedText struct {
	English   string
	Indonesia string
}

type BillDetail struct {
	BillerReferenceId string
	BillNo            string
	BillDescription   *LocalizedText
	BillSubCompany    string
	BillAmount        *Amount
	BillReferenceNo   string
	AdditionalInfo    map[string]interface{}
	Status            string
	Reason            *LocalizedText
}

type AdditionalInfoItem struct {
	Label *LocalizedText
	Value *LocalizedText
}

type OriginatorInfo struct {
	OriginatorCustomerNo   string
	OriginatorCustomerName string
	OriginatorBankCode     string
}
