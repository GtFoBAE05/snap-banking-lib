package bri

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"snap-banking-lib/internal/utils"
	"snap-banking-lib/model"
)

type briInternalAccountInquiryAdditionalInfo struct {
	DeviceId string `json:"deviceId,omitempty"`
	Channel  string `json:"channel,omitempty"`
}

type briInternalAccountInquiryRequest struct {
	BeneficiaryAccountNo string                                   `json:"beneficiaryAccountNo"`
	AdditionalInfo       *briInternalAccountInquiryAdditionalInfo `json:"additionalInfo,omitempty"`
}

type briInternalAccountInquiryResponse struct {
	ResponseCode             string `json:"responseCode"`
	ResponseMessage          string `json:"responseMessage"`
	ReferenceNo              string `json:"referenceNo"`
	BeneficiaryAccountName   string `json:"beneficiaryAccountName"`
	BeneficiaryAccountNo     string `json:"beneficiaryAccountNo"`
	BeneficiaryAccountStatus string `json:"beneficiaryAccountStatus"`
	BeneficiaryAccountType   string `json:"beneficiaryAccountType"`
	Currency                 string `json:"currency"`
}

func (a *BRIAdapter) GetInternalAccountInquiry(ctx context.Context, accessToken string, request *model.InternalAccountInquiryRequest) (*model.InternalAccountInquiryResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBRI.Name(), "GetInternalAccountInquiry")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointInternalAccountInquiry)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := briInternalAccountInquiryRequest{
		BeneficiaryAccountNo: request.BeneficiaryAccountNo,
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalInternalAccountInquiryRequest,
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
			Operation: model.OperationInternalAccountInquiryRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadInternalAccountInquiryResponse,
			Err:       err,
		}
	}

	var resp briInternalAccountInquiryResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalInternalAccountInquiryResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessInternalAccountInquiry {
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

	return briInternalAccountInquiryResponseToModel(resp), nil
}

func briInternalAccountInquiryResponseToModel(briResp briInternalAccountInquiryResponse) *model.InternalAccountInquiryResponse {
	return &model.InternalAccountInquiryResponse{
		ResponseCode:             briResp.ResponseCode,
		ResponseMessage:          briResp.ResponseMessage,
		ReferenceNo:              briResp.ReferenceNo,
		BeneficiaryAccountName:   briResp.BeneficiaryAccountName,
		BeneficiaryAccountNo:     briResp.BeneficiaryAccountNo,
		BeneficiaryAccountStatus: briResp.BeneficiaryAccountStatus,
		BeneficiaryAccountType:   briResp.BeneficiaryAccountType,
		Currency:                 briResp.Currency,
		Raw:                      briResp,
	}
}
