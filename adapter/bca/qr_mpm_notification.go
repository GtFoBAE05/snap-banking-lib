package bca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"snap-banking-lib/internal/utils"
	"snap-banking-lib/model"
	"time"
)

type bcaQRISNotificationMerchantInfo struct {
	TerminalId         string  `json:"terminalId"`
	MerchantId         string  `json:"merchantId"`
	City               string  `json:"city"`
	PostalCode         string  `json:"postalCode"`
	Country            string  `json:"country"`
	Email              *string `json:"email"`
	PaymentChannelName string  `json:"paymentChannelName"`
}

type bcaQRISNotificationAdditionalInfo struct {
	ReferenceNumber       string                           `json:"referenceNumber"`
	TransactionDate       string                           `json:"transactionDate"`
	ApprovalCode          string                           `json:"approvalCode"`
	PayerPhoneNumber      string                           `json:"payerPhoneNumber"`
	BatchNumber           string                           `json:"batchNumber"`
	ConvenienceFee        string                           `json:"convenienceFee"`
	IssuerReferenceNumber string                           `json:"issuerReferenceNumber"`
	PayerName             string                           `json:"payerName"`
	IssuerName            string                           `json:"issuerName"`
	AcquirerName          string                           `json:"acquirerName"`
	MerchantInfo          *bcaQRISNotificationMerchantInfo `json:"merchantInfo"`
}

type bcaQRISNotificationRequest struct {
	OriginalReferenceNo        string                             `json:"originalReferenceNo"`
	OriginalPartnerReferenceNo string                             `json:"originalPartnerReferenceNo"`
	LatestTransactionStatus    string                             `json:"latestTransactionStatus"`
	TransactionStatusDesc      string                             `json:"transactionStatusDesc"`
	CustomerNumber             string                             `json:"customerNumber"`
	AccountType                *string                            `json:"accountType"`
	DestinationNumber          string                             `json:"destinationNumber"`
	DestinationAccountName     string                             `json:"destinationAccountName"`
	Amount                     *model.Amount                      `json:"amount"`
	BankCode                   *string                            `json:"bankCode"`
	AdditionalInfo             *bcaQRISNotificationAdditionalInfo `json:"additionalInfo"`
}

type bcaQRISNotificationResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
}

func (a *BCAAdapter) HandleQRISNotification(ctx context.Context, accessToken string, request *model.QRISNotificationRequest) (*model.QRISNotificationResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "HandleQRISNotification")
	defer span.End()

	url, path, err := a.GetPartnerEndpoint(model.EndpointQRISNotification)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := bcaQRISNotificationRequest{
		OriginalReferenceNo:        request.OriginalReferenceNo,
		OriginalPartnerReferenceNo: request.OriginalPartnerReferenceNo,
		LatestTransactionStatus:    request.LatestTransactionStatus,
		TransactionStatusDesc:      request.TransactionStatusDesc,
		CustomerNumber:             request.CustomerNumber,
		AccountType:                request.AccountType,
		DestinationNumber:          request.DestinationNumber,
		DestinationAccountName:     request.DestinationAccountName,
		BankCode:                   request.BankCode,
	}

	if request.Amount != nil {
		requestBody.Amount = &model.Amount{
			Value:    request.Amount.Value,
			Currency: request.Amount.Currency,
		}
	}

	if request.AdditionalInfo != nil {
		info := request.AdditionalInfo
		requestBody.AdditionalInfo = &bcaQRISNotificationAdditionalInfo{
			ReferenceNumber:       info.ReferenceNumber,
			TransactionDate:       info.TransactionDate.Format(time.RFC3339),
			ApprovalCode:          info.ApprovalCode,
			PayerPhoneNumber:      info.PayerPhoneNumber,
			BatchNumber:           info.BatchNumber,
			ConvenienceFee:        info.ConvenienceFee,
			IssuerReferenceNumber: info.IssuerReferenceNumber,
			PayerName:             info.PayerName,
			IssuerName:            info.IssuerName,
			AcquirerName:          info.AcquirerName,
		}
		if info.MerchantInfo != nil {
			requestBody.AdditionalInfo.MerchantInfo = &bcaQRISNotificationMerchantInfo{
				TerminalId:         info.MerchantInfo.TerminalId,
				MerchantId:         info.MerchantInfo.MerchantId,
				City:               info.MerchantInfo.City,
				PostalCode:         info.MerchantInfo.PostalCode,
				Country:            info.MerchantInfo.Country,
				Email:              info.MerchantInfo.Email,
				PaymentChannelName: info.MerchantInfo.PaymentChannelName,
			}
		}
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalQRISNotificationRequest,
			Err:       err,
		}
	}

	timestamp := utils.ISO8601Timestamp()

	signature, err := a.GenerateServiceSignature(ctx, accessToken, "POST", path, timestamp, string(requestBodyJSON))
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
		"X-TIMESTAMP":   timestamp,
		"X-SIGNATURE":   signature,
		"X-PARTNER-ID":  request.PartnerId,
		"CHANNEL-ID":    request.ChannelId,
		"X-EXTERNAL-ID": request.ExternalId,
	}

	response, err := a.httpClient.Do(ctx, "POST", url, headers, requestBodyJSON)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationQRISNotificationRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadQRISNotificationResponse,
			Err:       err,
		}
	}

	var resp bcaQRISNotificationResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalQRISNotificationResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessQRISNotification {
		code, message := model.MapSNAPError(resp.ResponseCode, resp.ResponseMessage)
		apiErr := &model.APIError{
			Code:       code,
			Message:    message,
			HTTPStatus: response.StatusCode,
			RawCode:    resp.ResponseCode,
		}
		span.RecordError(apiErr)
		return nil, apiErr
	}

	return &model.QRISNotificationResponse{
		ResponseCode:    resp.ResponseCode,
		ResponseMessage: resp.ResponseMessage,
	}, nil
}
