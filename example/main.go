package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	snap "snap-banking-lib"
	"snap-banking-lib/adapter"
	"snap-banking-lib/model"
	"time"
)

type SlogLogger struct {
	log *slog.Logger
}

func (l *SlogLogger) Debug(msg string, args ...any) {
	l.log.Debug(msg, args...)
}

func (l *SlogLogger) Info(msg string, args ...any) {
	l.log.Info(msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...any) {
	l.log.Error(msg, args...)
}

var logger = &SlogLogger{
	log: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})),
}

func main() {
	config, err := model.LoadFromEnv(".env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client, err := snap.NewClient(*config,
		snap.WithLogger(logger),
		snap.WithDebug(),
		snap.WithCurlLog(),
		snap.WithRequestId("request_id", "request_id"),
		//snap.WithRetry(httpclient.RetryConfig{
		//	MaxAttempts: 3,
		//	Delay:       500 * time.Millisecond,
		//	MaxDelay:    5 * time.Second,
		//}),
		//snap.WithCircuitBreaker(httpclient.CircuitBreakerConfig{
		//	MaxRequests: 3,
		//	Interval:    10 * time.Second,
		//	Timeout:     30 * time.Second,
		//	Threshold:   5,
		//}),
	)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", fmt.Sprintf("%d", time.Now().UnixMilli()))

	logger.Info("==========BCA Example=====")

	bcaAdapter, err := client.GetAdapter(model.BankBCA)
	if err != nil {
		log.Fatalf("Failed to get BCA adapter: %v", err)
	}
	accessToken, err := bcaAdapter.GetAccessToken(ctx)
	if err != nil {
		logger.Error("get access token failed", "error", err)
		return
	}
	logger.Info("get access token success", "token", accessToken)
	bcaAccountInformation(ctx, bcaAdapter, accessToken)
	bcaInterbankTransfer(ctx, bcaAdapter, accessToken)
	bcaIntrabankTransfer(ctx, bcaAdapter, accessToken)

	logger.Info("==========BRI Example=====")

	briAdapter, err := client.GetAdapter(model.BankBRI)
	if err != nil {
		log.Fatalf("Failed to get BRI adapter: %v", err)
	}

	accessToken, err = briAdapter.GetAccessToken(ctx)
	if err != nil {
		logger.Error("get access token failed", "error", err)
		return
	}
	logger.Info("get access token success", "token", accessToken)

	briAccountInformation(ctx, briAdapter, accessToken)
	briInterbankTransfer(ctx, briAdapter, accessToken)
	briIntrabankTransfer(ctx, briAdapter, accessToken)
}

func bcaAccountInformation(ctx context.Context, bcaAdapter adapter.Adapter, accessToken *model.AccessToken) {
	inquiryResponse, err := bcaAdapter.GetBalanceInquiry(ctx, accessToken.AccessToken, &model.BalanceInquiryRequest{
		ChannelId:          "95051",
		PartnerId:          "KBBABCINDO",
		AccountNo:          "1234567890",
		PartnerReferenceNo: "2020102900000000000001",
	})
	if err != nil {
		logger.Error("get balance inquiry failed", "error", err)
	} else {
		logger.Info("get balance inquiry success", "response", inquiryResponse)
	}

	bankStatementResponse, err := bcaAdapter.GetBankStatement(ctx, accessToken.AccessToken, &model.BankStatementRequest{
		ChannelId:          "95051",
		PartnerId:          "KBBABCINDO",
		AccountNo:          "1234567890",
		PartnerReferenceNo: "2020102900000000000001",
		FromDateTime:       time.Now(),
		ToDateTime:         time.Now(),
	})
	if err != nil {
		logger.Error("get bank statement failed", "error", err)
	} else {
		logger.Info("get bank statement success", "response", bankStatementResponse)
	}
}

func bcaInterbankTransfer(ctx context.Context, bcaAdapter adapter.Adapter, accessToken *model.AccessToken) {
	internalAccountInquiryResponse, err := bcaAdapter.GetInternalAccountInquiry(ctx, accessToken.AccessToken, &model.InternalAccountInquiryRequest{
		ChannelId:            "95051",
		PartnerId:            "KBBABCINDO",
		ExternalId:           "28910000006578499987546738976812",
		Origin:               "www.hostname.com",
		PartnerReferenceNo:   "202010290000000001",
		BeneficiaryAccountNo: "8010001575",
	})
	if err != nil {
		logger.Error("get internal account inquiry failed", "error", err)
	} else {
		logger.Info("get internal account inquiry success", "response", internalAccountInquiryResponse)
	}

	interbankTransferResponse, err := bcaAdapter.CreateInterbankTransfer(ctx, accessToken.AccessToken, &model.InterbankTransferRequest{
		ChannelId:              "95051",
		PartnerId:              "KBBABCINDO",
		ExternalId:             "28910000006578499987546738976812",
		PartnerReferenceNo:     "2020102900000000000001",
		BeneficiaryAccountName: "Yories Yolanda",
		BeneficiaryAccountNo:   "888801000157508",
		BeneficiaryBankCode:    "BRINDIJA",
		BeneficiaryEmail:       "yories.yolanda@work.bri.co.id",
		SourceAccountNo:        "0123456789",
		TransactionDate:        time.Now(),
		Amount: &model.Amount{
			Value:    "10000.00",
			Currency: "IDR",
		},
		AdditionalInfo: &model.InterbankTransferAdditionalInfo{
			TransferType: "2",
			PurposeCode:  "02",
		},
	})
	if err != nil {
		logger.Error("create interbank transfer failed", "error", err)
	} else {
		logger.Info("create interbank transfer success", "response", interbankTransferResponse)
	}
}

func bcaIntrabankTransfer(ctx context.Context, bcaAdapter adapter.Adapter, accessToken *model.AccessToken) {
	externalAccountInquiryResponse, err := bcaAdapter.GetExternalAccountInquiry(ctx, accessToken.AccessToken, &model.ExternalAccountInquiryRequest{
		ChannelId:            "95051",
		PartnerId:            "KBBABCINDO",
		ExternalId:           "28910000006578499987546738976812",
		Origin:               "www.hostname.com",
		BeneficiaryBankCode:  "BRINIDJA",
		PartnerReferenceNo:   "2020102900000000000001",
		BeneficiaryAccountNo: "888801000157508",
		AdditionalInfo: &model.ExternalAccountInquiryAdditionalInfo{
			InquiryService:  "2",
			SourceAccountNo: "0123456789",
			Amount: &model.Amount{
				Value:    "10000.00",
				Currency: "IDR",
			},
			PurposeCode: "02",
		},
	})
	if err != nil {
		logger.Error("get external account inquiry failed", "error", err)
	} else {
		logger.Info("get external account inquiry success", "response", externalAccountInquiryResponse)
	}

	intrabankTransferResponse, err := bcaAdapter.CreateIntrabankTransfer(ctx, accessToken.AccessToken, &model.IntrabankTransferRequest{
		ChannelId:            "95051",
		PartnerId:            "KBBABCINDO",
		PartnerReferenceNo:   "2020102900000000000001",
		BeneficiaryAccountNo: "888801000157508",
		BeneficiaryEmail:     "yories.yolanda@work.bri.co.id",
		Remark:               "remark test",
		SourceAccountNo:      "888801000157508",
		TransactionDate:      time.Now(),
		Amount: &model.Amount{
			Value:    "10000.00",
			Currency: "IDR",
		},
		AdditionalInfo: &model.IntrabankTransferAdditionalInfo{
			EconomicActivity:   "Biaya Hidup Pihak Asing",
			TransactionPurpose: "01",
		},
	})
	if err != nil {
		logger.Error("create intrabank transfer failed", "error", err)
	} else {
		logger.Info("create intrabank transfer success", "response", intrabankTransferResponse)
	}
}

func briAccountInformation(ctx context.Context, briAdapter adapter.Adapter, accessToken *model.AccessToken) {
	// Balance Inquiry
	inquiryResponse, err := briAdapter.GetBalanceInquiry(ctx, accessToken.AccessToken, &model.BalanceInquiryRequest{
		ChannelId:  "123",
		PartnerId:  "123",
		ExternalId: "123",
		AccountNo:  "111231271284153",
	})
	if err != nil {
		logger.Error("get balance inquiry failed", "error", err)
	} else {
		logger.Info("get balance inquiry success", "response", inquiryResponse)
	}

	// Bank Statement
	fromTime, _ := time.Parse(time.RFC3339, "2024-03-08T10:41:45+07:00")
	toTime, _ := time.Parse(time.RFC3339, "2024-03-08T11:41:45+07:00")

	bankStatementResponse, err := briAdapter.GetBankStatement(ctx, accessToken.AccessToken, &model.BankStatementRequest{
		ChannelId:    "123",
		PartnerId:    "123",
		ExternalId:   "123",
		AccountNo:    "234567891012349",
		FromDateTime: fromTime,
		ToDateTime:   toTime,
	})
	if err != nil {
		logger.Error("get bank statement failed", "error", err)
	} else {
		logger.Info("get bank statement success", "response", bankStatementResponse)
	}
}

func briInterbankTransfer(ctx context.Context, briAdapter adapter.Adapter, accessToken *model.AccessToken) {
	externalAccountInquiryResponse, err := briAdapter.GetExternalAccountInquiry(ctx, accessToken.AccessToken, &model.ExternalAccountInquiryRequest{
		ChannelId:            "123",
		PartnerId:            "123",
		ExternalId:           "123",
		BeneficiaryBankCode:  "002",
		BeneficiaryAccountNo: "888801000157508",
		AdditionalInfo: &model.ExternalAccountInquiryAdditionalInfo{
			InquiryService: "16",
			DeviceId:       "12345679237",
			Channel:        "mobilephone",
		},
	})
	if err != nil {
		logger.Error("get external account inquiry failed", "error", err)
	} else {
		logger.Info("get external account inquiry success", "response", externalAccountInquiryResponse)
	}

	interbankTransferResponse, err := briAdapter.CreateInterbankTransfer(ctx, accessToken.AccessToken, &model.InterbankTransferRequest{
		ChannelId:              "123",
		PartnerId:              "123",
		ExternalId:             "123",
		PartnerReferenceNo:     "20211130000000001",
		BeneficiaryAccountName: "Dummy",
		BeneficiaryAccountNo:   "888801000134789",
		BeneficiaryAddress:     "Lorem City",
		BeneficiaryBankCode:    "002",
		BeneficiaryBankName:    "Bank Lorem Ipsum",
		BeneficiaryEmail:       "dummy.email@domain.com",
		CustomerReference:      "10052023",
		SourceAccountNo:        "988901000987654",
		TransactionDate:        time.Now(),
		Amount: &model.Amount{
			Value:    "10000.00",
			Currency: "IDR",
		},
		AdditionalInfo: &model.InterbankTransferAdditionalInfo{
			ServiceCode: "18",
			DeviceId:    "98765432101",
			Channel:     "mobilephone",
		},
		OriginatorInfos: []model.OriginatorInfo{
			{
				OriginatorCustomerNo:   "99901000004567",
				OriginatorCustomerName: "John Doe",
				OriginatorBankCode:     "003",
			},
		},
	})
	if err != nil {
		logger.Error("create interbank transfer failed", "error", err)
	} else {
		logger.Info("create interbank transfer success", "response", interbankTransferResponse)
	}

	transactionStatusResponse, err := briAdapter.GetTransactionStatusInquiry(ctx, accessToken.AccessToken, &model.TransactionStatusInquiryRequest{
		ChannelId:                  "123",
		PartnerId:                  "123",
		ExternalId:                 "123",
		OriginalPartnerReferenceNo: "20211130000000001",
		ServiceCode:                "18",
		TransactionDate:            time.Now(),
	})
	if err != nil {
		logger.Error("get transaction status inquiry failed", "error", err)
	} else {
		logger.Info("get transaction status inquiry success", "response", transactionStatusResponse)
	}
}

func briIntrabankTransfer(ctx context.Context, briAdapter adapter.Adapter, accessToken *model.AccessToken) {
	// Internal Account Inquiry
	internalAccountInquiryResponse, err := briAdapter.GetInternalAccountInquiry(ctx, accessToken.AccessToken, &model.InternalAccountInquiryRequest{
		ChannelId:            "123",
		PartnerId:            "123",
		ExternalId:           "123",
		BeneficiaryAccountNo: "888801000157508",
	})
	if err != nil {
		logger.Error("get internal account inquiry failed", "error", err)
	} else {
		logger.Info("get internal account inquiry success", "response", internalAccountInquiryResponse)
	}

	// Intrabank Transfer
	intrabankTransferResponse, err := briAdapter.CreateIntrabankTransfer(ctx, accessToken.AccessToken, &model.IntrabankTransferRequest{
		ChannelId:            "123",
		PartnerId:            "123",
		ExternalId:           "123",
		PartnerReferenceNo:   "2021112500000000000001",
		BeneficiaryAccountNo: "888801000157508",
		CustomerReference:    "10052031",
		FeeType:              "BEN",
		Remark:               "remark test",
		SourceAccountNo:      "888801000157610",
		TransactionDate:      time.Now(),
		Amount: &model.Amount{
			Value:    "100000.00",
			Currency: "IDR",
		},
		OriginatorInfos: []model.OriginatorInfo{
			{
				OriginatorCustomerNo:   "99901000003301",
				OriginatorCustomerName: "Iin",
				OriginatorBankCode:     "002",
			},
		},
		AdditionalInfo: &model.IntrabankTransferAdditionalInfo{
			DeviceId: "1234579237",
			Channel:  "mobilephone",
		},
	})
	if err != nil {
		logger.Error("create intrabank transfer failed", "error", err)
	} else {
		logger.Info("create intrabank transfer success", "response", intrabankTransferResponse)
	}

	transactionStatusResponse, err := briAdapter.GetTransactionStatusInquiry(ctx, accessToken.AccessToken, &model.TransactionStatusInquiryRequest{
		ChannelId:                  "123",
		PartnerId:                  "123",
		ExternalId:                 "123",
		OriginalPartnerReferenceNo: "100520138",
		ServiceCode:                "17",
		TransactionDate:            time.Now(),
	})
	if err != nil {
		logger.Error("get transaction status inquiry failed", "error", err)
	} else {
		logger.Info("get transaction status inquiry success", "response", transactionStatusResponse)
	}
}
