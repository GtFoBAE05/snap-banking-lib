package bri

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"snap-banking-lib/internal/utils"
	"snap-banking-lib/model"
)

type briVirtualAccountIntrabankInquiryRequest struct {
	PartnerServiceId string `json:"partnerServiceId"`
	CustomerNo       string `json:"customerNo"`
	VirtualAccountNo string `json:"virtualAccountNo"`
}

type briVirtualAccountIntrabankInquiryAdditionalInfo struct {
	Description string `json:"description,omitempty"`
}

type briVirtualAccountIntrabankInquiryDataResponse struct {
	PartnerServiceId   string                                           `json:"partnerServiceId"`
	CustomerNo         string                                           `json:"customerNo"`
	VirtualAccountNo   string                                           `json:"virtualAccountNo"`
	VirtualAccountName string                                           `json:"virtualAccountName"`
	TotalAmount        *model.Amount                                    `json:"totalAmount"`
	AdditionalInfo     *briVirtualAccountIntrabankInquiryAdditionalInfo `json:"additionalInfo"`
}

type briVirtualAccountIntrabankInquiryResponse struct {
	ResponseCode       string                                         `json:"responseCode"`
	ResponseMessage    string                                         `json:"responseMessage"`
	VirtualAccountData *briVirtualAccountIntrabankInquiryDataResponse `json:"virtualAccountData"`
}

func (a *BRIAdapter) GetVirtualAccountIntrabankInquiry(ctx context.Context, accessToken string, request *model.VirtualAccountIntrabankInquiryRequest) (*model.VirtualAccountIntrabankInquiryResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBRI.Name(), "GetVirtualAccountIntrabankInquiry")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointVirtualAccountIntrabankInquiry)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := briVirtualAccountIntrabankInquiryRequest{
		PartnerServiceId: request.PartnerServiceId,
		CustomerNo:       request.CustomerNo,
		VirtualAccountNo: request.VirtualAccountNo,
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalVirtualAccountIntrabankInquiryRequest,
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
			Operation: model.OperationVirtualAccountIntrabankInquiryRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadVirtualAccountIntrabankInquiryResponse,
			Err:       err,
		}
	}

	var resp briVirtualAccountIntrabankInquiryResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalVirtualAccountIntrabankInquiryResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessVirtualAccountIntrabankInquiry {
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

	return briVirtualAccountIntrabankInquiryResponseToModel(resp), nil
}

func briVirtualAccountIntrabankInquiryResponseToModel(briResp briVirtualAccountIntrabankInquiryResponse) *model.VirtualAccountIntrabankInquiryResponse {
	modelResp := &model.VirtualAccountIntrabankInquiryResponse{
		ResponseCode:    briResp.ResponseCode,
		ResponseMessage: briResp.ResponseMessage,
		Raw:             briResp,
	}

	if briResp.VirtualAccountData == nil {
		return modelResp
	}

	data := briResp.VirtualAccountData
	modelResp.VirtualAccountData = &model.VirtualAccountIntrabankInquiryData{
		PartnerServiceId:   data.PartnerServiceId,
		CustomerNo:         data.CustomerNo,
		VirtualAccountNo:   data.VirtualAccountNo,
		VirtualAccountName: data.VirtualAccountName,
	}

	if data.TotalAmount != nil {
		modelResp.VirtualAccountData.TotalAmount = &model.Amount{
			Value:    data.TotalAmount.Value,
			Currency: data.TotalAmount.Currency,
		}
	}

	return modelResp
}
