package model

import "fmt"

// Operation
const (
	OperationGetEndpoint = "get_endpoint"

	// Authentication
	OperationMarshalAccessTokenRequest    = "marshal_access_token_request"
	OperationUnmarshalAccessTokenResponse = "unmarshal_access_token_response"
	OperationAccessTokenRequest           = "access_token_request"
	OperationReadAccessTokenResponse      = "read_access_token_response"
	OperationParseExpiresIn               = "parse_expires_in"
	OperationLoadPrivateKey               = "load_private_key"
	OperationSignRequest                  = "sign_request"

	// Balance Inquiry operations
	OperationMarshalBalanceInquiryRequest    = "marshal_balance_inquiry_request"
	OperationUnmarshalBalanceInquiryResponse = "unmarshal_balance_inquiry_response"
	OperationBalanceInquiryRequest           = "balance_inquiry_request"
	OperationReadBalanceInquiryResponse      = "read_balance_inquiry_response"
	OperationGenerateBalanceInquirySignature = "generate_balance_inquiry_signature"

	// Bank Statement operations
	OperationMarshalBankStatementRequest    = "marshal_bank_statement_request"
	OperationUnmarshalBankStatementResponse = "unmarshal_bank_statement_response"
	OperationBankStatementRequest           = "bank_statement_request"
	OperationReadBankStatementResponse      = "read_bank_statement_response"

	// Virtual Account Inquiry operations
	OperationMarshalVirtualAccountInquiryRequest    = "marshal_virtual_account_inquiry_request"
	OperationUnmarshalVirtualAccountInquiryResponse = "unmarshal_virtual_account_inquiry_response"
	OperationVirtualAccountInquiryRequest           = "virtual_account_inquiry_request"
	OperationReadVirtualAccountInquiryResponse      = "read_virtual_account_inquiry_response"

	// Virtual Account Payment operations
	OperationMarshalVirtualAccountPaymentRequest    = "marshal_virtual_account_payment_request"
	OperationUnmarshalVirtualAccountPaymentResponse = "unmarshal_virtual_account_payment_response"
	OperationVirtualAccountPaymentRequest           = "virtual_account_payment_request"
	OperationReadVirtualAccountPaymentResponse      = "read_virtual_account_payment_response"

	// Virtual Account Inquiry Status operations
	OperationMarshalVirtualAccountInquiryStatusRequest    = "marshal_virtual_account_inquiry_status_request"
	OperationUnmarshalVirtualAccountInquiryStatusResponse = "unmarshal_virtual_account_inquiry_status_response"
	OperationVirtualAccountInquiryStatusRequest           = "virtual_account_inquiry_status_request"
	OperationReadVirtualAccountInquiryStatusResponse      = "read_virtual_account_inquiry_status_response"

	// Virtual Account Intrabank Inquiry operations
	OperationMarshalVirtualAccountIntrabankInquiryRequest    = "marshal_virtual_account_intrabank_inquiry_request"
	OperationUnmarshalVirtualAccountIntrabankInquiryResponse = "unmarshal_virtual_account_intrabank_inquiry_response"
	OperationVirtualAccountIntrabankInquiryRequest           = "virtual_account_intrabank_inquiry_request"
	OperationReadVirtualAccountIntrabankInquiryResponse      = "read_virtual_account_intrabank_inquiry_response"

	// Virtual Account Intrabank Payment Notification operations
	OperationMarshalVirtualAccountIntrabankPaymentNotificationRequest    = "marshal_virtual_account_intrabank_payment_notification_request"
	OperationUnmarshalVirtualAccountIntrabankPaymentNotificationResponse = "unmarshal_virtual_account_intrabank_payment_notification_response"
	OperationVirtualAccountIntrabankPaymentNotificationRequest           = "virtual_account_intrabank_payment_notification_request"
	OperationReadVirtualAccountIntrabankPaymentNotificationResponse      = "read_virtual_account_intrabank_payment_notification_response"

	// Virtual Account Intrabank Payment operations
	OperationMarshalVirtualAccountIntrabankPaymentRequest    = "marshal_virtual_account_intrabank_payment_request"
	OperationUnmarshalVirtualAccountIntrabankPaymentResponse = "unmarshal_virtual_account_intrabank_payment_response"
	OperationVirtualAccountIntrabankPaymentRequest           = "virtual_account_intrabank_payment_request"
	OperationReadVirtualAccountIntrabankPaymentResponse      = "read_virtual_account_intrabank_payment_response"

	// QR MPM Generate operations
	OperationMarshalQRMPMGenerateRequest    = "marshal_qr_mpm_generate_request"
	OperationUnmarshalQRMPMGenerateResponse = "unmarshal_qr_mpm_generate_response"
	OperationQRMPMGenerateRequest           = "qr_mpm_generate_request"
	OperationReadQRMPMGenerateResponse      = "read_qr_mpm_generate_response"

	// QR MPM Inquiry operations
	OperationMarshalQRMPMInquiryRequest    = "marshal_qr_mpm_inquiry_request"
	OperationUnmarshalQRMPMInquiryResponse = "unmarshal_qr_mpm_inquiry_response"
	OperationQRMPMInquiryRequest           = "qr_mpm_inquiry_request"
	OperationReadQRMPMInquiryResponse      = "read_qr_mpm_inquiry_response"

	// QR MPM Refund operations
	OperationMarshalQRMPMRefundRequest    = "marshal_qr_mpm_refund_request"
	OperationUnmarshalQRMPMRefundResponse = "unmarshal_qr_mpm_refund_response"
	OperationQRMPMRefundRequest           = "qr_mpm_refund_request"
	OperationReadQRMPMRefundResponse      = "read_qr_mpm_refund_response"

	// QRIS Notification operations
	OperationMarshalQRISNotificationRequest    = "marshal_qris_notification_request"
	OperationUnmarshalQRISNotificationResponse = "unmarshal_qris_notification_response"
	OperationQRISNotificationRequest           = "qris_notification_request"
	OperationReadQRISNotificationResponse      = "read_qris_notification_response"

	// External Account Inquiry operations
	OperationMarshalExternalAccountInquiryRequest    = "marshal_external_account_inquiry_request"
	OperationUnmarshalExternalAccountInquiryResponse = "unmarshal_external_account_inquiry_response"
	OperationExternalAccountInquiryRequest           = "external_account_inquiry_request"
	OperationReadExternalAccountInquiryResponse      = "read_external_account_inquiry_response"

	// Interbank Transfer operations
	OperationMarshalInterbankTransferRequest    = "marshal_interbank_transfer_request"
	OperationUnmarshalInterbankTransferResponse = "unmarshal_interbank_transfer_response"
	OperationInterbankTransferRequest           = "interbank_transfer_request"
	OperationReadInterbankTransferResponse      = "read_interbank_transfer_response"

	// Internal Account Inquiry operations
	OperationMarshalInternalAccountInquiryRequest    = "marshal_internal_account_inquiry_request"
	OperationUnmarshalInternalAccountInquiryResponse = "unmarshal_internal_account_inquiry_response"
	OperationInternalAccountInquiryRequest           = "internal_account_inquiry_request"
	OperationReadInternalAccountInquiryResponse      = "read_internal_account_inquiry_response"

	// Intrabank Transfer operations
	OperationMarshalIntrabankTransferRequest    = "marshal_intrabank_transfer_request"
	OperationUnmarshalIntrabankTransferResponse = "unmarshal_intrabank_transfer_response"
	OperationIntrabankTransferRequest           = "intrabank_transfer_request"
	OperationReadIntrabankTransferResponse      = "read_intrabank_transfer_response"

	// Transaction Status Inquiry operations
	OperationMarshalTransactionStatusInquiryRequest    = "marshal_transaction_status_inquiry_request"
	OperationUnmarshalTransactionStatusInquiryResponse = "unmarshal_transaction_status_inquiry_response"
	OperationTransactionStatusInquiryRequest           = "transaction_status_inquiry_request"
	OperationReadTransactionStatusInquiryResponse      = "read_transaction_status_inquiry_response"
)

// API Error Const
const (
	ErrInvalidFieldFormat         = "invalid_field_format"
	ErrUnauthorized               = "unauthorized"
	ErrInvalidTimestampFormat     = "invalid_timestamp_format"
	ErrMissingField               = "missing_field"
	ErrTimeout                    = "timeout"
	ErrInvalidSignature           = "invalid_signature"
	ErrTooManyRequests            = "too_many_requests"
	ErrConflict                   = "conflict"
	ErrPaidBill                   = "paid_bill"
	ErrInvalidBill                = "invalid_bill"
	ErrInconsistentRequest        = "inconsistent_request"
	ErrTransactionNotFound        = "transaction_not_found"
	ErrExceedsTransactionLimit    = "exceeds_transaction_limit"
	ErrActivityCountLimitExceeded = "activity_count_limit_exceeded"
	ErrInvalidMerchant            = "invalid_merchant"
	ErrTransactionExpired         = "transaction_expired"
	ErrDoNotHonor                 = "do_not_honor"
	ErrTransactionNotPermitted    = "transaction_not_permitted"
	ErrInvalidAmount              = "invalid_amount"
	ErrMethodNotAllowed           = "method_not_allowed"
	ErrInvalidRouting             = "invalid_routing"
	ErrInsufficientFunds          = "insufficient_funds"
	ErrBankNotSupported           = "bank_not_supported"
	ErrInactiveAccount            = "inactive_account"
	ErrUnknown                    = "unknown_error"
	ErrNotSupported               = "not_supported"
)

func MapSNAPError(responseCode, responseMessage string) (code, message string) {
	message = responseMessage

	if len(responseCode) < 5 {
		return ErrUnknown, message
	}

	httpStatus := responseCode[:3]
	caseCode := responseCode[len(responseCode)-2:]

	switch httpStatus {
	case "400":
		switch caseCode {
		case "02":
			code = ErrMissingField
		default: // 00=BadRequest, 01=InvalidFieldFormat
			code = ErrInvalidFieldFormat
		}

	case "401":
		switch caseCode {
		case "01", "02", "03", "04":
			code = ErrUnauthorized
		default:
			if containsSignature(responseMessage) {
				code = ErrInvalidSignature
			} else {
				code = ErrUnauthorized
			}
		}

	case "403":
		switch caseCode {
		case "00":
			code = ErrTransactionExpired
		case "01", "06":
			code = ErrUnauthorized
		case "02":
			code = ErrExceedsTransactionLimit
		case "04":
			code = ErrActivityCountLimitExceeded
		case "05":
			code = ErrDoNotHonor
		case "09":
			code = ErrInactiveAccount
		case "14":
			code = ErrInsufficientFunds
		case "15":
			code = ErrTransactionNotPermitted
		case "18":
			code = ErrInactiveAccount
		default:
			code = ErrUnauthorized
		}

	case "404":
		switch caseCode {
		case "01":
			code = ErrTransactionNotFound
		case "02":
			code = ErrInvalidRouting
		case "03":
			code = ErrBankNotSupported
		case "08":
			code = ErrInvalidMerchant
		case "11":
			code = ErrInvalidFieldFormat
		case "12":
			code = ErrInvalidBill
		case "13":
			code = ErrInvalidAmount
		case "14":
			code = ErrPaidBill
		case "18":
			code = ErrInconsistentRequest
		default:
			code = ErrTransactionNotFound
		}

	case "405":
		code = ErrMethodNotAllowed

	case "409":
		code = ErrConflict

	case "429":
		code = ErrTooManyRequests

	case "500", "504":
		code = ErrTimeout

	default:
		code = ErrUnknown
	}

	return
}

func containsSignature(msg string) bool {
	target := "Signature"
	if len(msg) < len(target) {
		return false
	}
	for i := 0; i <= len(msg)-len(target); i++ {
		if msg[i:i+len(target)] == target {
			return true
		}
	}
	return false
}

type APIError struct {
	Code       string
	Message    string
	HTTPStatus int
	RawCode    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

type ClientError struct {
	Operation string
	Err       error
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("client error during %s: %v", e.Operation, e.Err)
}

func (e *ClientError) Unwrap() error {
	return e.Err
}

type NetworkError struct {
	Operation string
	Err       error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error during %s: %v", e.Operation, e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}
