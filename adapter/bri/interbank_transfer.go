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

type briOriginatorInfo struct {
	OriginatorCustomerNo   string `json:"originatorCustomerNo"`
	OriginatorCustomerName string `json:"originatorCustomerName"`
	OriginatorBankCode     string `json:"originatorBankCode"`
}

type briInterbankTransferAdditionalInfo struct {
	ServiceCode          string `json:"serviceCode"`
	DeviceId             string `json:"deviceId,omitempty"`
	Channel              string `json:"channel,omitempty"`
	ReferenceNo          string `json:"referenceNo,omitempty"`
	ExternalId           string `json:"externalId,omitempty"`
	SenderIdentityNumber string `json:"senderIdentityNumber,omitempty"`
	PaymentInfo          string `json:"paymentInfo,omitempty"`
	SenderType           string `json:"senderType,omitempty"`
	SenderResidentStatus string `json:"senderResidentStatus,omitempty"`
	SenderTownName       string `json:"senderTownName,omitempty"`
	IsRdn                string `json:"isRdn,omitempty"`
}

type briInterbankTransferRequest struct {
	PartnerReferenceNo     string                              `json:"partnerReferenceNo"`
	Amount                 *model.Amount                       `json:"amount"`
	BeneficiaryAccountName string                              `json:"beneficiaryAccountName,omitempty"`
	BeneficiaryAccountNo   string                              `json:"beneficiaryAccountNo"`
	BeneficiaryAddress     string                              `json:"beneficiaryAddress,omitempty"`
	BeneficiaryBankCode    string                              `json:"beneficiaryBankCode,omitempty"`
	BeneficiaryBankName    string                              `json:"beneficiaryBankName,omitempty"`
	BeneficiaryEmail       string                              `json:"beneficiaryEmail,omitempty"`
	CustomerReference      string                              `json:"customerReference,omitempty"`
	SourceAccountNo        string                              `json:"sourceAccountNo"`
	TransactionDate        string                              `json:"transactionDate"`
	AdditionalInfo         *briInterbankTransferAdditionalInfo `json:"additionalInfo,omitempty"`
	OriginatorInfos        []briOriginatorInfo                 `json:"originatorInfos,omitempty"`
}

type briInterbankTransferResponse struct {
	ResponseCode         string              `json:"responseCode"`
	ResponseMessage      string              `json:"responseMessage"`
	ReferenceNo          string              `json:"referenceNo"`
	PartnerReferenceNo   string              `json:"partnerReferenceNo"`
	Amount               *model.Amount       `json:"amount"`
	BeneficiaryAccountNo string              `json:"beneficiaryAccountNo"`
	BeneficiaryBankCode  string              `json:"beneficiaryBankCode"`
	SourceAccountNo      string              `json:"sourceAccountNo"`
	OriginatorInfos      []briOriginatorInfo `json:"originatorInfos"`
}

func (a *BRIAdapter) CreateInterbankTransfer(ctx context.Context, accessToken string, request *model.InterbankTransferRequest) (*model.InterbankTransferResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBRI.Name(), "CreateInterbankTransfer")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointInterbankTransfer)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := briInterbankTransferRequest{
		PartnerReferenceNo:     request.PartnerReferenceNo,
		BeneficiaryAccountName: request.BeneficiaryAccountName,
		BeneficiaryAccountNo:   request.BeneficiaryAccountNo,
		BeneficiaryAddress:     request.BeneficiaryAddress,
		BeneficiaryBankCode:    request.BeneficiaryBankCode,
		BeneficiaryBankName:    request.BeneficiaryBankName,
		BeneficiaryEmail:       request.BeneficiaryEmail,
		CustomerReference:      request.CustomerReference,
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
		requestBody.AdditionalInfo = &briInterbankTransferAdditionalInfo{
			ServiceCode:          request.AdditionalInfo.ServiceCode,
			DeviceId:             request.AdditionalInfo.DeviceId,
			Channel:              request.AdditionalInfo.Channel,
			ReferenceNo:          request.AdditionalInfo.ReferenceNo,
			ExternalId:           request.AdditionalInfo.ExternalId,
			SenderIdentityNumber: request.AdditionalInfo.SenderIdentityNumber,
			PaymentInfo:          request.AdditionalInfo.PaymentInfo,
			SenderType:           request.AdditionalInfo.SenderType,
			SenderResidentStatus: request.AdditionalInfo.SenderResidentStatus,
			SenderTownName:       request.AdditionalInfo.SenderTownName,
			IsRdn:                request.AdditionalInfo.IsRdn,
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

	var resp briInterbankTransferResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalInterbankTransferResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessInterbankTransfer && resp.ResponseCode != SuccessInterbankTransferBIFAST {
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

	return briInterbankTransferResponseToModel(resp), nil
}

func briInterbankTransferResponseToModel(briResp briInterbankTransferResponse) *model.InterbankTransferResponse {
	modelResp := &model.InterbankTransferResponse{
		ResponseCode:         briResp.ResponseCode,
		ResponseMessage:      briResp.ResponseMessage,
		ReferenceNo:          briResp.ReferenceNo,
		PartnerReferenceNo:   briResp.PartnerReferenceNo,
		BeneficiaryAccountNo: briResp.BeneficiaryAccountNo,
		BeneficiaryBankCode:  briResp.BeneficiaryBankCode,
		SourceAccountNo:      briResp.SourceAccountNo,
		Raw:                  briResp,
	}

	if briResp.Amount != nil {
		modelResp.Amount = &model.Amount{
			Value:    briResp.Amount.Value,
			Currency: briResp.Amount.Currency,
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
