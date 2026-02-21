package adapter

import (
	"context"
	"snap-banking-lib/model"
)

type Adapter interface {
	GetBankCode() model.BankCode
	GetBankName() string
	GetBaseURL() string
	GetPartnerBaseURL() string
	GetEndpoint(key model.EndpointKey) (url string, path string, err error)
	GetPartnerEndpoint(key model.EndpointKey) (url string, path string, err error)

	// Signature
	GenerateTokenSignature(ctx context.Context, timestamp string) (string, error)
	GenerateServiceSignature(ctx context.Context, accessToken, method, path, timestamp, body string) (string, error)

	// Authentication
	GetAccessToken(ctx context.Context) (*model.AccessToken, error)

	// Account Operations
	GetBalanceInquiry(ctx context.Context, accessToken string, request *model.BalanceInquiryRequest) (*model.BalanceInquiryResponse, error)
	GetBankStatement(ctx context.Context, accessToken string, request *model.BankStatementRequest) (*model.BankStatementResponse, error)

	// Virtual Account
	GetVirtualAccountInquiry(ctx context.Context, accessToken string, request *model.VirtualAccountInquiryRequest) (*model.VirtualAccountInquiryResponse, error)
	GetVirtualAccountInquiryStatus(ctx context.Context, accessToken string, request *model.VirtualAccountInquiryStatusRequest) (*model.VirtualAccountInquiryStatusResponse, error)
	GetVirtualAccountPayment(ctx context.Context, accessToken string, request *model.VirtualAccountPaymentRequest) (*model.VirtualAccountPaymentResponse, error)

	// Virtual Account Intrabank
	GetVirtualAccountIntrabankInquiry(ctx context.Context, accessToken string, request *model.VirtualAccountIntrabankInquiryRequest) (*model.VirtualAccountIntrabankInquiryResponse, error)
	GetVirtualAccountIntrabankPaymentNotification(ctx context.Context, accessToken string, request *model.VirtualAccountIntrabankPaymentNotificationRequest) (*model.VirtualAccountIntrabankPaymentNotificationResponse, error)
	GetVirtualAccountIntrabankPayment(ctx context.Context, accessToken string, request *model.VirtualAccountIntrabankPaymentRequest) (*model.VirtualAccountIntrabankPaymentResponse, error)

	// QR MPM
	GenerateQRMPM(ctx context.Context, accessToken string, request *model.QRMPMGenerateRequest) (*model.QRMPMGenerateResponse, error)
	GetQRMPMInquiry(ctx context.Context, accessToken string, request *model.QRMPMInquiryRequest) (*model.QRMPMInquiryResponse, error)
	RefundQRMPM(ctx context.Context, accessToken string, request *model.QRMPMRefundRequest) (*model.QRMPMRefundResponse, error)
	HandleQRISNotification(ctx context.Context, accessToken string, request *model.QRISNotificationRequest) (*model.QRISNotificationResponse, error)

	// Transfer
	GetExternalAccountInquiry(ctx context.Context, accessToken string, request *model.ExternalAccountInquiryRequest) (*model.ExternalAccountInquiryResponse, error)
	GetInternalAccountInquiry(ctx context.Context, accessToken string, request *model.InternalAccountInquiryRequest) (*model.InternalAccountInquiryResponse, error)
	CreateInterbankTransfer(ctx context.Context, accessToken string, request *model.InterbankTransferRequest) (*model.InterbankTransferResponse, error)
	CreateIntrabankTransfer(ctx context.Context, accessToken string, request *model.IntrabankTransferRequest) (*model.IntrabankTransferResponse, error)
	GetTransactionStatusInquiry(ctx context.Context, accessToken string, request *model.TransactionStatusInquiryRequest) (*model.TransactionStatusInquiryResponse, error)
}
