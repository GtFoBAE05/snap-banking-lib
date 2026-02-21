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

type briVirtualAccountIntrabankPaymentRequest struct {
	PartnerServiceId   string        `json:"partnerServiceId"`
	CustomerNo         string        `json:"customerNo"`
	VirtualAccountNo   string        `json:"virtualAccountNo"`
	VirtualAccountName string        `json:"virtualAccountName"`
	SourceAccountNo    string        `json:"sourceAccountNo"`
	PartnerReferenceNo string        `json:"partnerReferenceNo"`
	PaidAmount         *model.Amount `json:"paidAmount"`
	TrxDateTime        string        `json:"trxDateTime"`
}

type briVirtualAccountIntrabankPaymentDataResponse struct {
	PartnerServiceId   string        `json:"partnerServiceId"`
	CustomerNo         string        `json:"customerNo"`
	VirtualAccountNo   string        `json:"virtualAccountNo"`
	VirtualAccountName string        `json:"virtualAccountName"`
	PartnerReferenceNo string        `json:"partnerReferenceNo"`
	PaymentRequestId   string        `json:"paymentRequestId"`
	PaidAmount         *model.Amount `json:"paidAmount"`
	TrxDateTime        string        `json:"trxDateTime"`
}

type briVirtualAccountIntrabankPaymentResponse struct {
	ResponseCode       string                                         `json:"responseCode"`
	ResponseMessage    string                                         `json:"responseMessage"`
	VirtualAccountData *briVirtualAccountIntrabankPaymentDataResponse `json:"virtualAccountData"`
}

func (a *BRIAdapter) GetVirtualAccountIntrabankPayment(ctx context.Context, accessToken string, request *model.VirtualAccountIntrabankPaymentRequest) (*model.VirtualAccountIntrabankPaymentResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBRI.Name(), "GetVirtualAccountIntrabankPayment")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointVirtualAccountIntrabankPayment)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := briVirtualAccountIntrabankPaymentRequest{
		PartnerServiceId:   request.PartnerServiceId,
		CustomerNo:         request.CustomerNo,
		VirtualAccountNo:   request.VirtualAccountNo,
		VirtualAccountName: request.VirtualAccountName,
		SourceAccountNo:    request.SourceAccountNo,
		PartnerReferenceNo: request.PartnerReferenceNo,
		TrxDateTime:        request.TrxDateTime.Format(time.RFC3339),
	}

	if request.PaidAmount != nil {
		requestBody.PaidAmount = &model.Amount{
			Value:    request.PaidAmount.Value,
			Currency: request.PaidAmount.Currency,
		}
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalVirtualAccountIntrabankPaymentRequest,
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
			Operation: model.OperationVirtualAccountIntrabankPaymentRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadVirtualAccountIntrabankPaymentResponse,
			Err:       err,
		}
	}

	var resp briVirtualAccountIntrabankPaymentResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalVirtualAccountIntrabankPaymentResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessVirtualAccountIntrabankPayment {
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

	return briVirtualAccountIntrabankPaymentResponseToModel(resp), nil
}

func briVirtualAccountIntrabankPaymentResponseToModel(briResp briVirtualAccountIntrabankPaymentResponse) *model.VirtualAccountIntrabankPaymentResponse {
	modelResp := &model.VirtualAccountIntrabankPaymentResponse{
		ResponseCode:    briResp.ResponseCode,
		ResponseMessage: briResp.ResponseMessage,
		Raw:             briResp,
	}

	if briResp.VirtualAccountData == nil {
		return modelResp
	}

	data := briResp.VirtualAccountData
	modelResp.VirtualAccountData = &model.VirtualAccountIntrabankPaymentData{
		PartnerServiceId:   data.PartnerServiceId,
		CustomerNo:         data.CustomerNo,
		VirtualAccountNo:   data.VirtualAccountNo,
		VirtualAccountName: data.VirtualAccountName,
		PartnerReferenceNo: data.PartnerReferenceNo,
		PaymentRequestId:   data.PaymentRequestId,
		TrxDateTime:        data.TrxDateTime,
	}

	if data.PaidAmount != nil {
		modelResp.VirtualAccountData.PaidAmount = &model.Amount{
			Value:    data.PaidAmount.Value,
			Currency: data.PaidAmount.Currency,
		}
	}

	return modelResp
}
