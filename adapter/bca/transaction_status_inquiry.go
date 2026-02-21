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

type bcaTransactionStatusInquiryRequest struct {
	OriginalPartnerReferenceNo string `json:"originalPartnerReferenceNo"`
	OriginalExternalId         string `json:"originalExternalId"`
	ServiceCode                string `json:"serviceCode"`
	TransactionDate            string `json:"transactionDate"`
}

type bcaTransactionStatusInquiryResponse struct {
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

func (a *BCAAdapter) GetTransactionStatusInquiry(ctx context.Context, accessToken string, request *model.TransactionStatusInquiryRequest) (*model.TransactionStatusInquiryResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "GetTransactionStatusInquiry")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointTransactionStatusInquiry)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := bcaTransactionStatusInquiryRequest{
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

	var resp bcaTransactionStatusInquiryResponse
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

	return bcaTransactionStatusInquiryResponseToModel(resp), nil
}

func bcaTransactionStatusInquiryResponseToModel(bcaResp bcaTransactionStatusInquiryResponse) *model.TransactionStatusInquiryResponse {
	modelResp := &model.TransactionStatusInquiryResponse{
		ResponseCode:               bcaResp.ResponseCode,
		ResponseMessage:            bcaResp.ResponseMessage,
		OriginalReferenceNo:        bcaResp.OriginalReferenceNo,
		OriginalPartnerReferenceNo: bcaResp.OriginalPartnerReferenceNo,
		OriginalExternalId:         bcaResp.OriginalExternalId,
		ServiceCode:                bcaResp.ServiceCode,
		TransactionDate:            bcaResp.TransactionDate,
		BeneficiaryAccountNo:       bcaResp.BeneficiaryAccountNo,
		BeneficiaryBankCode:        bcaResp.BeneficiaryBankCode,
		ReferenceNumber:            bcaResp.ReferenceNumber,
		SourceAccountNo:            bcaResp.SourceAccountNo,
		LatestTransactionStatus:    bcaResp.LatestTransactionStatus,
		TransactionStatusDesc:      bcaResp.TransactionStatusDesc,
		Raw:                        bcaResp,
	}

	if bcaResp.Amount != nil {
		modelResp.Amount = &model.Amount{
			Value:    bcaResp.Amount.Value,
			Currency: bcaResp.Amount.Currency,
		}
	}

	return modelResp
}
