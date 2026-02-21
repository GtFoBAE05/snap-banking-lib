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

type bcaInterbankTransferAdditionalInfo struct {
	TransferType string `json:"transferType"`
	PurposeCode  string `json:"purposeCode"`
}

type bcaInterbankTransferRequest struct {
	PartnerReferenceNo     string                              `json:"partnerReferenceNo"`
	Amount                 *model.Amount                       `json:"amount"`
	BeneficiaryAccountName string                              `json:"beneficiaryAccountName"`
	BeneficiaryAccountNo   string                              `json:"beneficiaryAccountNo"`
	BeneficiaryBankCode    string                              `json:"beneficiaryBankCode"`
	BeneficiaryEmail       string                              `json:"beneficiaryEmail"`
	SourceAccountNo        string                              `json:"sourceAccountNo"`
	TransactionDate        string                              `json:"transactionDate"`
	AdditionalInfo         *bcaInterbankTransferAdditionalInfo `json:"additionalInfo"`
}

type bcaInterbankTransferResponse struct {
	ResponseCode         string        `json:"responseCode"`
	ResponseMessage      string        `json:"responseMessage"`
	ReferenceNo          string        `json:"referenceNo"`
	PartnerReferenceNo   string        `json:"partnerReferenceNo"`
	Amount               *model.Amount `json:"amount"`
	BeneficiaryAccountNo string        `json:"beneficiaryAccountNo"`
	BeneficiaryBankCode  string        `json:"beneficiaryBankCode"`
	SourceAccountNo      string        `json:"sourceAccountNo"`
}

func (a *BCAAdapter) CreateInterbankTransfer(ctx context.Context, accessToken string, request *model.InterbankTransferRequest) (*model.InterbankTransferResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "CreateInterbankTransfer")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointInterbankTransfer)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := bcaInterbankTransferRequest{
		PartnerReferenceNo:     request.PartnerReferenceNo,
		BeneficiaryAccountName: request.BeneficiaryAccountName,
		BeneficiaryAccountNo:   request.BeneficiaryAccountNo,
		BeneficiaryBankCode:    request.BeneficiaryBankCode,
		BeneficiaryEmail:       request.BeneficiaryEmail,
		SourceAccountNo:        request.SourceAccountNo,
		TransactionDate:        request.TransactionDate.Format(time.RFC3339),
	}

	if request.Amount != nil {
		requestBody.Amount = &model.Amount{
			Value:    request.Amount.Value,
			Currency: request.Amount.Currency,
		}
	}

	if request.AdditionalInfo != nil {
		requestBody.AdditionalInfo = &bcaInterbankTransferAdditionalInfo{
			TransferType: request.AdditionalInfo.TransferType,
			PurposeCode:  request.AdditionalInfo.PurposeCode,
		}
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalInterbankTransferRequest,
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
			Operation: model.OperationInterbankTransferRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadInterbankTransferResponse,
			Err:       err,
		}
	}

	var resp bcaInterbankTransferResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalInterbankTransferResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessInterbankTransfer {
		code, message := model.MapSNAPError(resp.ResponseCode, resp.ResponseMessage)
		span.RecordError(err)
		return nil, &model.APIError{
			Code:       code,
			Message:    message,
			HTTPStatus: response.StatusCode,
			RawCode:    resp.ResponseCode,
		}
	}

	return bcaInterbankTransferResponseToModel(resp), nil
}

func bcaInterbankTransferResponseToModel(bcaResp bcaInterbankTransferResponse) *model.InterbankTransferResponse {
	modelResp := &model.InterbankTransferResponse{
		ResponseCode:         bcaResp.ResponseCode,
		ResponseMessage:      bcaResp.ResponseMessage,
		ReferenceNo:          bcaResp.ReferenceNo,
		PartnerReferenceNo:   bcaResp.PartnerReferenceNo,
		BeneficiaryAccountNo: bcaResp.BeneficiaryAccountNo,
		BeneficiaryBankCode:  bcaResp.BeneficiaryBankCode,
		SourceAccountNo:      bcaResp.SourceAccountNo,
		Raw:                  bcaResp,
	}

	if bcaResp.Amount != nil {
		modelResp.Amount = &model.Amount{
			Value:    bcaResp.Amount.Value,
			Currency: bcaResp.Amount.Currency,
		}
	}

	return modelResp
}
