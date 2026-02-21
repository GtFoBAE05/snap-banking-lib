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

type bcaVirtualAccountIntrabankPaymentNotificationRequest struct {
	VirtualAccountNo   string            `json:"virtualAccountNo"`
	PartnerReferenceNo string            `json:"partnerReferenceNo"`
	TrxDateTime        string            `json:"trxDateTime"`
	PaymentStatus      string            `json:"paymentStatus"`
	PaymentFlagReason  *bcaLocalizedText `json:"paymentFlagReason"`
}

type bcaVirtualAccountIntrabankPaymentNotificationData struct {
	VirtualAccountNo   string `json:"virtualAccountNo"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
}

type bcaVirtualAccountIntrabankPaymentNotificationResponse struct {
	ResponseCode       string                                             `json:"responseCode"`
	ResponseMessage    string                                             `json:"responseMessage"`
	VirtualAccountData *bcaVirtualAccountIntrabankPaymentNotificationData `json:"virtualAccountData"`
}

func (a *BCAAdapter) GetVirtualAccountIntrabankPaymentNotification(ctx context.Context, accessToken string, request *model.VirtualAccountIntrabankPaymentNotificationRequest) (*model.VirtualAccountIntrabankPaymentNotificationResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "GetVirtualAccountIntrabankPaymentNotification")
	defer span.End()

	url, path, err := a.GetPartnerEndpoint(model.EndpointVirtualAccountIntrabankPaymentNotification)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := bcaVirtualAccountIntrabankPaymentNotificationRequest{
		VirtualAccountNo:   request.VirtualAccountNo,
		PartnerReferenceNo: request.PartnerReferenceNo,
		TrxDateTime:        request.TrxDateTime.Format(time.RFC3339),
		PaymentStatus:      request.PaymentStatus,
	}

	if request.PaymentFlagReason != nil {
		requestBody.PaymentFlagReason = &bcaLocalizedText{
			English:   request.PaymentFlagReason.English,
			Indonesia: request.PaymentFlagReason.Indonesia,
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

	var resp bcaVirtualAccountIntrabankPaymentNotificationResponse
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

	return bcaVirtualAccountIntrabankPaymentNotificationResponseToModel(resp), nil
}

func bcaVirtualAccountIntrabankPaymentNotificationResponseToModel(bcaResp bcaVirtualAccountIntrabankPaymentNotificationResponse) *model.VirtualAccountIntrabankPaymentNotificationResponse {
	modelResp := &model.VirtualAccountIntrabankPaymentNotificationResponse{
		ResponseCode:    bcaResp.ResponseCode,
		ResponseMessage: bcaResp.ResponseMessage,
		Raw:             bcaResp,
	}

	if bcaResp.VirtualAccountData == nil {
		return modelResp
	}

	modelResp.VirtualAccountData = &model.VirtualAccountIntrabankPaymentNotificationData{
		VirtualAccountNo:   bcaResp.VirtualAccountData.VirtualAccountNo,
		PartnerReferenceNo: bcaResp.VirtualAccountData.PartnerReferenceNo,
	}

	return modelResp
}
