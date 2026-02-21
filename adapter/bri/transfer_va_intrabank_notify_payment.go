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

type briVirtualAccountIntrabankPaymentNotificationAdditionalInfo struct {
	IdApp         string `json:"idApp,omitempty"`
	PassApp       string `json:"passApp,omitempty"`
	PaymentAmount string `json:"paymentAmount"`
	TerminalId    string `json:"terminalId,omitempty"`
	BankId        string `json:"bankId,omitempty"`
}

type briVirtualAccountIntrabankPaymentNotificationRequest struct {
	PartnerServiceId string                                                       `json:"partnerServiceId"`
	CustomerNo       string                                                       `json:"customerNo"`
	VirtualAccountNo string                                                       `json:"virtualAccountNo"`
	PaymentRequestId string                                                       `json:"paymentRequestId"`
	TrxDateTime      string                                                       `json:"trxDateTime"`
	AdditionalInfo   *briVirtualAccountIntrabankPaymentNotificationAdditionalInfo `json:"additionalInfo,omitempty"`
}

type briVirtualAccountIntrabankPaymentNotificationDataResponse struct {
	PartnerServiceId string `json:"partnerServiceId"`
	CustomerNo       string `json:"customerNo"`
	VirtualAccountNo string `json:"virtualAccountNo"`
	InquiryRequestId string `json:"inquiryRequestId"`
	PaymentRequestId string `json:"paymentRequestId"`
	TrxDateTime      string `json:"trxDateTime"`
	PaymentStatus    string `json:"paymentStatus"`
}

type briVirtualAccountIntrabankPaymentNotificationResponse struct {
	ResponseCode       string                                                     `json:"responseCode"`
	ResponseMessage    string                                                     `json:"responseMessage"`
	VirtualAccountData *briVirtualAccountIntrabankPaymentNotificationDataResponse `json:"virtualAccountData"`
}

func (a *BRIAdapter) GetVirtualAccountIntrabankPaymentNotification(ctx context.Context, accessToken string, request *model.VirtualAccountIntrabankPaymentNotificationRequest) (*model.VirtualAccountIntrabankPaymentNotificationResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBRI.Name(), "GetVirtualAccountIntrabankPaymentNotification")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointVirtualAccountIntrabankPaymentNotification)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := briVirtualAccountIntrabankPaymentNotificationRequest{
		PartnerServiceId: request.PartnerServiceId,
		CustomerNo:       request.CustomerNo,
		VirtualAccountNo: request.VirtualAccountNo,
		PaymentRequestId: request.PaymentRequestId,
		TrxDateTime:      request.TrxDateTime.Format(time.RFC3339),
	}

	if request.AdditionalInfo != nil {
		requestBody.AdditionalInfo = &briVirtualAccountIntrabankPaymentNotificationAdditionalInfo{
			IdApp:         request.AdditionalInfo.IdApp,
			PassApp:       request.AdditionalInfo.PassApp,
			PaymentAmount: request.AdditionalInfo.PaymentAmount,
			TerminalId:    request.AdditionalInfo.TerminalId,
			BankId:        request.AdditionalInfo.BankId,
		}
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalVirtualAccountIntrabankPaymentNotificationRequest,
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
			Operation: model.OperationVirtualAccountIntrabankPaymentNotificationRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadVirtualAccountIntrabankPaymentNotificationResponse,
			Err:       err,
		}
	}

	var resp briVirtualAccountIntrabankPaymentNotificationResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalVirtualAccountIntrabankPaymentNotificationResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessVirtualAccountIntrabankPaymentNotification {
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

	return briVirtualAccountIntrabankPaymentNotificationResponseToModel(resp), nil
}

func briVirtualAccountIntrabankPaymentNotificationResponseToModel(briResp briVirtualAccountIntrabankPaymentNotificationResponse) *model.VirtualAccountIntrabankPaymentNotificationResponse {
	modelResp := &model.VirtualAccountIntrabankPaymentNotificationResponse{
		ResponseCode:    briResp.ResponseCode,
		ResponseMessage: briResp.ResponseMessage,
		Raw:             briResp,
	}

	if briResp.VirtualAccountData == nil {
		return modelResp
	}

	modelResp.VirtualAccountData = &model.VirtualAccountIntrabankPaymentNotificationData{
		PartnerServiceId: briResp.VirtualAccountData.PartnerServiceId,
		CustomerNo:       briResp.VirtualAccountData.CustomerNo,
		VirtualAccountNo: briResp.VirtualAccountData.VirtualAccountNo,
		InquiryRequestId: briResp.VirtualAccountData.InquiryRequestId,
		PaymentRequestId: briResp.VirtualAccountData.PaymentRequestId,
		TrxDateTime:      briResp.VirtualAccountData.TrxDateTime,
		PaymentStatus:    briResp.VirtualAccountData.PaymentStatus,
	}

	return modelResp
}
