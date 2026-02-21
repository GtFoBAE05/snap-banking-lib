package model

type Endpoint struct {
	Path string
}

type EndpointKey string

const (
	// Authentication
	EndpointAccessToken EndpointKey = "access_token"

	// Account Information
	EndpointBalanceInquiry EndpointKey = "balance_inquiry"
	EndpointBankStatement  EndpointKey = "bank_statement"

	// Virtual Account
	EndpointVirtualAccountInquiry       EndpointKey = "va_inquiry"
	EndpointVirtualAccountInquiryStatus EndpointKey = "va_inquiry_status"
	EndpointVirtualAccountPayment       EndpointKey = "va_payment"

	// Virtual Account Intrabank
	EndpointVirtualAccountIntrabankInquiry             EndpointKey = "va_intrabank_inquiry"
	EndpointVirtualAccountIntrabankPaymentNotification EndpointKey = "va_intrabank_payment_notification"
	EndpointVirtualAccountIntrabankPayment             EndpointKey = "va_intrabank_payment"

	// QR
	EndpointQRMPMGenerate    EndpointKey = "qr_mpm_generate"
	EndpointQRMPMInquiry     EndpointKey = "qr_mpm_inquiry"
	EndpointQRMPMRefund      EndpointKey = "qr_mpm_refund"
	EndpointQRISNotification EndpointKey = "qris_notification"

	// Transfer
	EndpointExternalAccountInquiry   EndpointKey = "external_account_inquiry"
	EndpointInterbankTransfer        EndpointKey = "interbank_transfer"
	EndpointInternalAccountInquiry   EndpointKey = "internal_account_inquiry"
	EndpointIntrabankTransfer        EndpointKey = "intrabank_transfer"
	EndpointTransactionStatusInquiry EndpointKey = "transaction_status_inquiry"
)

var DefaultBCAEndpoints = map[EndpointKey]Endpoint{
	EndpointAccessToken:    {Path: "/openapi/v1.0/access-token/b2b"},
	EndpointBalanceInquiry: {Path: "/openapi/v1.0/balance-inquiry"},
	EndpointBankStatement:  {Path: "/openapi/v1.0/bank-statement"},

	EndpointVirtualAccountInquiry:                      {Path: "/openapi/v1.0/transfer-va/inquiry"},
	EndpointVirtualAccountInquiryStatus:                {Path: "/openapi/v1.0/transfer-va/status"},
	EndpointVirtualAccountPayment:                      {Path: "/openapi/v1.0/transfer-va/payment"},
	EndpointVirtualAccountIntrabankInquiry:             {Path: "/openapi/v1.0/transfer-va/inquiry-intrabank"},
	EndpointVirtualAccountIntrabankPaymentNotification: {Path: "/openapi/v1.0/transfer-va/notify-payment-intrabank"},
	EndpointVirtualAccountIntrabankPayment:             {Path: "/openapi/v1.0/transfer-va/payment-intrabank"},

	EndpointQRMPMGenerate:    {Path: "/openapi/v1.0/qr/qr-mpm-generate"},
	EndpointQRMPMInquiry:     {Path: "/openapi/v1.0/qr/qr-mpm-query"},
	EndpointQRMPMRefund:      {Path: "/openapi/v1.0/qr/qr-mpm-refund"},
	EndpointQRISNotification: {Path: "/openapi/v1.0/qr-mpm-notify"},

	EndpointExternalAccountInquiry:   {Path: "/openapi/v2.0/account-inquiry-external"},
	EndpointInterbankTransfer:        {Path: "/openapi/v2.0/transfer-interbank"},
	EndpointInternalAccountInquiry:   {Path: "/openapi/v1.0/account-inquiry-internal"},
	EndpointIntrabankTransfer:        {Path: "/openapi/v1.0/transfer-intrabank"},
	EndpointTransactionStatusInquiry: {Path: "/openapi/v1.0/transfer/status"},
}

var DefaultBRIEndpoints = map[EndpointKey]Endpoint{
	EndpointAccessToken:    {Path: "/snap/v1.0/access-token/b2b"},
	EndpointBalanceInquiry: {Path: "/snap/v1.0/balance-inquiry"},
	EndpointBankStatement:  {Path: "/snap/v1.1/bank-statement"},

	EndpointVirtualAccountInquiry:                      {Path: "/snap/v1.0/transfer-va/inquiry"},
	EndpointVirtualAccountInquiryStatus:                {Path: "/snap/v1.0/transfer-va/status"},
	EndpointVirtualAccountPayment:                      {Path: "/snap/v1.0/transfer-va/payment"},
	EndpointVirtualAccountIntrabankInquiry:             {Path: "/snap/v1.1/transfer-va/inquiry-intrabank"},
	EndpointVirtualAccountIntrabankPaymentNotification: {Path: "/snap/v1.0/transfer-va/notify-payment-intrabank"},
	EndpointVirtualAccountIntrabankPayment:             {Path: "/snap/v1.1/transfer-va/payment-intrabank"},

	EndpointExternalAccountInquiry:   {Path: "/interbank/snap/v1.1/account-inquiry-external"},
	EndpointInterbankTransfer:        {Path: "/interbank/snap/v1.0/transfer-interbank"},
	EndpointInternalAccountInquiry:   {Path: "/intrabank/snap/v2.0/account-inquiry-internal"},
	EndpointIntrabankTransfer:        {Path: "/intrabank/snap/v2.0/transfer-intrabank"},
	EndpointTransactionStatusInquiry: {Path: "/intrabank/snap/v1.0/transfer/status"},
}
