package bri

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"snap-banking-lib/internal/utils"
	"snap-banking-lib/model"
)

type briExternalAccountInquiryAdditionalInfo struct {
	ServiceCode string `json:"serviceCode"`
	DeviceId    string `json:"deviceId,omitempty"`
	Channel     string `json:"channel,omitempty"`
}

type briExternalAccountInquiryRequest struct {
	BeneficiaryBankCode  string                                   `json:"beneficiaryBankCode"`
	BeneficiaryAccountNo string                                   `json:"beneficiaryAccountNo"`
	AdditionalInfo       *briExternalAccountInquiryAdditionalInfo `json:"additionalInfo,omitempty"`
}

type briExternalAccountInquiryResponse struct {
	ResponseCode           string `json:"responseCode"`
	ResponseMessage        string `json:"responseMessage"`
	ReferenceNo            string `json:"referenceNo"`
	BeneficiaryAccountName string `json:"beneficiaryAccountName"`
	BeneficiaryAccountNo   string `json:"beneficiaryAccountNo"`
	BeneficiaryBankCode    string `json:"beneficiaryBankCode"`
	BeneficiaryBankName    string `json:"beneficiaryBankName"`
	Currency               string `json:"currency"`
}

func (a *BRIAdapter) GetExternalAccountInquiry(ctx context.Context, accessToken string, request *model.ExternalAccountInquiryRequest) (*model.ExternalAccountInquiryResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBRI.Name(), "GetExternalAccountInquiry")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointExternalAccountInquiry)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := briExternalAccountInquiryRequest{
		BeneficiaryBankCode:  request.BeneficiaryBankCode,
		BeneficiaryAccountNo: request.BeneficiaryAccountNo,
	}

	if request.AdditionalInfo != nil {
		requestBody.AdditionalInfo = &briExternalAccountInquiryAdditionalInfo{
			ServiceCode: request.AdditionalInfo.ServiceCode,
			DeviceId:    request.AdditionalInfo.DeviceId,
			Channel:     request.AdditionalInfo.Channel,
		}
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalExternalAccountInquiryRequest,
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
			Operation: model.OperationExternalAccountInquiryRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadExternalAccountInquiryResponse,
			Err:       err,
		}
	}

	var resp briExternalAccountInquiryResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalExternalAccountInquiryResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessExternalAccountInquiry && resp.ResponseCode != SuccessExternalAccountInquiryBIFAST {
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

	return briExternalAccountInquiryResponseToModel(resp), nil
}

func briExternalAccountInquiryResponseToModel(briResp briExternalAccountInquiryResponse) *model.ExternalAccountInquiryResponse {
	return &model.ExternalAccountInquiryResponse{
		ResponseCode:           briResp.ResponseCode,
		ResponseMessage:        briResp.ResponseMessage,
		ReferenceNo:            briResp.ReferenceNo,
		BeneficiaryAccountName: briResp.BeneficiaryAccountName,
		BeneficiaryAccountNo:   briResp.BeneficiaryAccountNo,
		BeneficiaryBankCode:    briResp.BeneficiaryBankCode,
		BeneficiaryBankName:    briResp.BeneficiaryBankName,
		Currency:               briResp.Currency,
		Raw:                    briResp,
	}
}
