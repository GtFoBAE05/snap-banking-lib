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

type briIntrabankTransferAdditionalInfo struct {
	DeviceId string `json:"deviceId,omitempty"`
	Channel  string `json:"channel,omitempty"`
	IsRdn    string `json:"isRdn,omitempty"`
}

type briIntrabankTransferRequest struct {
	PartnerReferenceNo   string                              `json:"partnerReferenceNo"`
	Amount               *model.Amount                       `json:"amount"`
	BeneficiaryAccountNo string                              `json:"beneficiaryAccountNo"`
	CustomerReference    string                              `json:"customerReference,omitempty"`
	FeeType              string                              `json:"feeType,omitempty"`
	OriginatorInfos      []briOriginatorInfo                 `json:"originatorInfos,omitempty"`
	Remark               string                              `json:"remark"`
	SourceAccountNo      string                              `json:"sourceAccountNo"`
	TransactionDate      string                              `json:"transactionDate"`
	AdditionalInfo       *briIntrabankTransferAdditionalInfo `json:"additionalInfo,omitempty"`
}

type briIntrabankTransferResponse struct {
	ResponseCode         string                              `json:"responseCode"`
	ResponseMessage      string                              `json:"responseMessage"`
	ReferenceNo          string                              `json:"referenceNo"`
	PartnerReferenceNo   string                              `json:"partnerReferenceNo"`
	Amount               *model.Amount                       `json:"amount"`
	BeneficiaryAccountNo string                              `json:"beneficiaryAccountNo"`
	CustomerReference    string                              `json:"customerReference"`
	SourceAccountNo      string                              `json:"sourceAccountNo"`
	TransactionDate      string                              `json:"transactionDate"`
	OriginatorInfos      []briOriginatorInfo                 `json:"originatorInfos"`
	AdditionalInfo       *briIntrabankTransferAdditionalInfo `json:"additionalInfo"`
}

func (a *BRIAdapter) CreateIntrabankTransfer(ctx context.Context, accessToken string, request *model.IntrabankTransferRequest) (*model.IntrabankTransferResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBRI.Name(), "CreateIntrabankTransfer")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointIntrabankTransfer)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := briIntrabankTransferRequest{
		PartnerReferenceNo:   request.PartnerReferenceNo,
		BeneficiaryAccountNo: request.BeneficiaryAccountNo,
		CustomerReference:    request.CustomerReference,
		FeeType:              request.FeeType,
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
		requestBody.AdditionalInfo = &briIntrabankTransferAdditionalInfo{
			DeviceId: request.AdditionalInfo.DeviceId,
			Channel:  request.AdditionalInfo.Channel,
			IsRdn:    request.AdditionalInfo.IsRdn,
		}
	}

	for _, o := range request.OriginatorInfos {
		requestBody.OriginatorInfos = append(requestBody.OriginatorInfos, briOriginatorInfo{
			OriginatorCustomerNo:   o.OriginatorCustomerNo,
			OriginatorCustomerName: o.OriginatorCustomerName,
			OriginatorBankCode:     o.OriginatorBankCode,
		})
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

	var resp briIntrabankTransferResponse
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

	return briIntrabankTransferResponseToModel(resp), nil
}

func briIntrabankTransferResponseToModel(briResp briIntrabankTransferResponse) *model.IntrabankTransferResponse {
	modelResp := &model.IntrabankTransferResponse{
		ResponseCode:         briResp.ResponseCode,
		ResponseMessage:      briResp.ResponseMessage,
		ReferenceNo:          briResp.ReferenceNo,
		PartnerReferenceNo:   briResp.PartnerReferenceNo,
		BeneficiaryAccountNo: briResp.BeneficiaryAccountNo,
		CustomerReference:    briResp.CustomerReference,
		SourceAccountNo:      briResp.SourceAccountNo,
		TransactionDate:      briResp.TransactionDate,
		Raw:                  briResp,
	}

	if briResp.Amount != nil {
		modelResp.Amount = &model.Amount{
			Value:    briResp.Amount.Value,
			Currency: briResp.Amount.Currency,
		}
	}

	if briResp.AdditionalInfo != nil {
		modelResp.AdditionalInfo = &model.IntrabankTransferAdditionalInfo{
			DeviceId: briResp.AdditionalInfo.DeviceId,
			Channel:  briResp.AdditionalInfo.Channel,
			IsRdn:    briResp.AdditionalInfo.IsRdn,
		}
	}

	for _, o := range briResp.OriginatorInfos {
		modelResp.OriginatorInfos = append(modelResp.OriginatorInfos, model.OriginatorInfo{
			OriginatorCustomerNo:   o.OriginatorCustomerNo,
			OriginatorCustomerName: o.OriginatorCustomerName,
			OriginatorBankCode:     o.OriginatorBankCode,
		})
	}

	return modelResp
}
