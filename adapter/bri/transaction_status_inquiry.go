package bri

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"snap-banking-lib/internal/utils"
	"snap-banking-lib/model"
	"time"
)

type briTransactionStatusInquiryRequest struct {
	OriginalPartnerReferenceNo string `json:"originalPartnerReferenceNo"`
	OriginalExternalId         string `json:"originalExternalId"`
	ServiceCode                string `json:"serviceCode"`
	TransactionDate            string `json:"transactionDate"`
}

type briTransactionStatusInquiryResponse struct {
	ResponseCode               string        `json:"responseCode"`
	ResponseMessage            string        `json:"responseMessage"`
	OriginalReferenceNo        string        `json:"originalReferenceNo"`
	OriginalPartnerReferenceNo string        `json:"originalPartnerReferenceNo"`
	OriginalExternalId         string        `json:"originalExternalId"`
	ServiceCode                string        `json:"serviceCode"`
	TransactionDate            string        `json:"transactionDate"`
	Amount                     *model.Amount `json:"amount"`
	BeneficiaryAccountNo       string        `json:"beneficiaryAccountNo"`
	BeneficiaryBankCode        string        `json:"beneficiaryBankCode"`
	ReferenceNumber            string        `json:"referenceNumber"`
	SourceAccountNo            string        `json:"sourceAccountNo"`
	LatestTransactionStatus    string        `json:"latestTransactionStatus"`
	TransactionStatusDesc      string        `json:"transactionStatusDesc"`
}

func (a *BRIAdapter) GetTransactionStatusInquiry(ctx context.Context, accessToken string, request *model.TransactionStatusInquiryRequest) (*model.TransactionStatusInquiryResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBRI.Name(), "GetTransactionStatusInquiry")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointTransactionStatusInquiry)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := briTransactionStatusInquiryRequest{
		OriginalPartnerReferenceNo: request.OriginalPartnerReferenceNo,
		OriginalExternalId:         request.OriginalExternalId,
		ServiceCode:                request.ServiceCode,
		TransactionDate:            request.TransactionDate.Format(time.RFC3339),
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalTransactionStatusInquiryRequest,
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
			Operation: model.OperationTransactionStatusInquiryRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadTransactionStatusInquiryResponse,
			Err:       err,
		}
	}

	var resp briTransactionStatusInquiryResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalTransactionStatusInquiryResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessTransactionStatusInquiry {
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

	return briTransactionStatusInquiryResponseToModel(resp), nil
}

func briTransactionStatusInquiryResponseToModel(briResp briTransactionStatusInquiryResponse) *model.TransactionStatusInquiryResponse {
	modelResp := &model.TransactionStatusInquiryResponse{
		ResponseCode:               briResp.ResponseCode,
		ResponseMessage:            briResp.ResponseMessage,
		OriginalReferenceNo:        briResp.OriginalReferenceNo,
		OriginalPartnerReferenceNo: briResp.OriginalPartnerReferenceNo,
		OriginalExternalId:         briResp.OriginalExternalId,
		ServiceCode:                briResp.ServiceCode,
		TransactionDate:            briResp.TransactionDate,
		BeneficiaryAccountNo:       briResp.BeneficiaryAccountNo,
		BeneficiaryBankCode:        briResp.BeneficiaryBankCode,
		ReferenceNumber:            briResp.ReferenceNumber,
		SourceAccountNo:            briResp.SourceAccountNo,
		LatestTransactionStatus:    briResp.LatestTransactionStatus,
		TransactionStatusDesc:      briResp.TransactionStatusDesc,
		Raw:                        briResp,
	}

	if briResp.Amount != nil {
		modelResp.Amount = &model.Amount{
			Value:    briResp.Amount.Value,
			Currency: briResp.Amount.Currency,
		}
	}

	return modelResp
}
