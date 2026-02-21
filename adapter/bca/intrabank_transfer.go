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

type bcaIntrabankTransferAdditionalInfo struct {
	EconomicActivity   string `json:"economicActivity"`
	TransactionPurpose string `json:"transactionPurpose"`
}

type bcaIntrabankTransferRequest struct {
	PartnerReferenceNo   string                              `json:"partnerReferenceNo"`
	Amount               *model.Amount                       `json:"amount"`
	BeneficiaryAccountNo string                              `json:"beneficiaryAccountNo"`
	BeneficiaryEmail     string                              `json:"beneficiaryEmail"`
	Remark               string                              `json:"remark"`
	SourceAccountNo      string                              `json:"sourceAccountNo"`
	TransactionDate      string                              `json:"transactionDate"`
	AdditionalInfo       *bcaIntrabankTransferAdditionalInfo `json:"additionalInfo"`
}

type bcaIntrabankTransferResponse struct {
	ResponseCode         string                              `json:"responseCode"`
	ResponseMessage      string                              `json:"responseMessage"`
	ReferenceNo          string                              `json:"referenceNo"`
	PartnerReferenceNo   string                              `json:"partnerReferenceNo"`
	Amount               *model.Amount                       `json:"amount"`
	BeneficiaryAccountNo string                              `json:"beneficiaryAccountNo"`
	SourceAccountNo      string                              `json:"sourceAccountNo"`
	TransactionDate      string                              `json:"transactionDate"`
	AdditionalInfo       *bcaIntrabankTransferAdditionalInfo `json:"additionalInfo"`
}

func (a *BCAAdapter) CreateIntrabankTransfer(ctx context.Context, accessToken string, request *model.IntrabankTransferRequest) (*model.IntrabankTransferResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "CreateIntrabankTransfer")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointIntrabankTransfer)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := bcaIntrabankTransferRequest{
		PartnerReferenceNo:   request.PartnerReferenceNo,
		BeneficiaryAccountNo: request.BeneficiaryAccountNo,
		BeneficiaryEmail:     request.BeneficiaryEmail,
		Remark:               request.Remark,
		SourceAccountNo:      request.SourceAccountNo,
		TransactionDate:      request.TransactionDate.Format(time.RFC3339),
	}

	if request.Amount != nil {
		requestBody.Amount = &model.Amount{
			Value:    request.Amount.Value,
			Currency: request.Amount.Currency,
		}
	}

	if request.AdditionalInfo != nil {
		requestBody.AdditionalInfo = &bcaIntrabankTransferAdditionalInfo{
			EconomicActivity:   request.AdditionalInfo.EconomicActivity,
			TransactionPurpose: request.AdditionalInfo.TransactionPurpose,
		}
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalIntrabankTransferRequest,
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
			Operation: model.OperationIntrabankTransferRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadIntrabankTransferResponse,
			Err:       err,
		}
	}

	var resp bcaIntrabankTransferResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalIntrabankTransferResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessIntrabankTransfer {
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

	return bcaIntrabankTransferResponseToModel(resp), nil
}

func bcaIntrabankTransferResponseToModel(bcaResp bcaIntrabankTransferResponse) *model.IntrabankTransferResponse {
	modelResp := &model.IntrabankTransferResponse{
		ResponseCode:         bcaResp.ResponseCode,
		ResponseMessage:      bcaResp.ResponseMessage,
		ReferenceNo:          bcaResp.ReferenceNo,
		PartnerReferenceNo:   bcaResp.PartnerReferenceNo,
		BeneficiaryAccountNo: bcaResp.BeneficiaryAccountNo,
		SourceAccountNo:      bcaResp.SourceAccountNo,
		TransactionDate:      bcaResp.TransactionDate,
		Raw:                  bcaResp,
	}

	if bcaResp.Amount != nil {
		modelResp.Amount = &model.Amount{
			Value:    bcaResp.Amount.Value,
			Currency: bcaResp.Amount.Currency,
		}
	}

	if bcaResp.AdditionalInfo != nil {
		modelResp.AdditionalInfo = &model.IntrabankTransferAdditionalInfo{
			EconomicActivity:   bcaResp.AdditionalInfo.EconomicActivity,
			TransactionPurpose: bcaResp.AdditionalInfo.TransactionPurpose,
		}
	}

	return modelResp
}
