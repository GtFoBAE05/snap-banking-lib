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

type bcaQRMPMRefundAdditionalInfo struct {
	TerminalId          string `json:"terminalId"`
	TransactionDate     string `json:"transactionDate"`
	PartnerMerchantType string `json:"partnerMerchantType"`
	IssuerName          string `json:"issuerName"`
}

type bcaQRMPMRefundRequest struct {
	MerchantId                 string                        `json:"merchantId"`
	OriginalPartnerReferenceNo string                        `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo        string                        `json:"originalReferenceNo"`
	PartnerRefundNo            string                        `json:"partnerRefundNo"`
	RefundAmount               *model.Amount                 `json:"refundAmount"`
	AdditionalInfo             *bcaQRMPMRefundAdditionalInfo `json:"additionalInfo"`
}

type bcaQRMPMRefundResponseAdditionalInfo struct {
	MerchantId      string        `json:"merchantId"`
	TerminalId      string        `json:"terminalId"`
	ReferenceNumber string        `json:"referenceNumber"`
	AvailableAmount *model.Amount `json:"availableAmount"`
	RefundCounter   string        `json:"refundCounter"`
}

type bcaQRMPMRefundResponse struct {
	ResponseCode               string                                `json:"responseCode"`
	ResponseMessage            string                                `json:"responseMessage"`
	OriginalPartnerReferenceNo string                                `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo        string                                `json:"originalReferenceNo"`
	OriginalExternalId         string                                `json:"originalExternalId"`
	RefundNo                   string                                `json:"refundNo"`
	PartnerRefundNo            string                                `json:"partnerRefundNo"`
	RefundAmount               *model.Amount                         `json:"refundAmount"`
	RefundTime                 string                                `json:"refundTime"`
	AdditionalInfo             *bcaQRMPMRefundResponseAdditionalInfo `json:"additionalInfo"`
}

func (a *BCAAdapter) RefundQRMPM(ctx context.Context, accessToken string, request *model.QRMPMRefundRequest) (*model.QRMPMRefundResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "RefundQRMPM")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointQRMPMRefund)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := bcaQRMPMRefundRequest{
		MerchantId:                 request.MerchantId,
		OriginalPartnerReferenceNo: request.OriginalPartnerReferenceNo,
		OriginalReferenceNo:        request.OriginalReferenceNo,
		PartnerRefundNo:            request.PartnerRefundNo,
	}

	if request.RefundAmount != nil {
		requestBody.RefundAmount = &model.Amount{
			Value:    request.RefundAmount.Value,
			Currency: request.RefundAmount.Currency,
		}
	}

	if request.AdditionalInfo != nil {
		requestBody.AdditionalInfo = &bcaQRMPMRefundAdditionalInfo{
			TerminalId:          request.AdditionalInfo.TerminalId,
			TransactionDate:     request.AdditionalInfo.TransactionDate.Format(time.RFC3339),
			PartnerMerchantType: request.AdditionalInfo.PartnerMerchantType,
			IssuerName:          request.AdditionalInfo.IssuerName,
		}
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalQRMPMRefundRequest,
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
			Operation: model.OperationQRMPMRefundRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadQRMPMRefundResponse,
			Err:       err,
		}
	}

	var resp bcaQRMPMRefundResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalQRMPMRefundResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessQRMPMRefund {
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

	return bcaQRMPMRefundResponseToModel(resp), nil
}

func bcaQRMPMRefundResponseToModel(bcaResp bcaQRMPMRefundResponse) *model.QRMPMRefundResponse {
	modelResp := &model.QRMPMRefundResponse{
		ResponseCode:               bcaResp.ResponseCode,
		ResponseMessage:            bcaResp.ResponseMessage,
		OriginalPartnerReferenceNo: bcaResp.OriginalPartnerReferenceNo,
		OriginalReferenceNo:        bcaResp.OriginalReferenceNo,
		OriginalExternalId:         bcaResp.OriginalExternalId,
		RefundNo:                   bcaResp.RefundNo,
		PartnerRefundNo:            bcaResp.PartnerRefundNo,
		RefundTime:                 bcaResp.RefundTime,
		Raw:                        bcaResp,
	}

	if bcaResp.RefundAmount != nil {
		modelResp.RefundAmount = &model.Amount{
			Value:    bcaResp.RefundAmount.Value,
			Currency: bcaResp.RefundAmount.Currency,
		}
	}

	if bcaResp.AdditionalInfo != nil {
		info := bcaResp.AdditionalInfo
		modelResp.AdditionalInfo = &model.QRMPMRefundResponseAdditionalInfo{
			MerchantId:      info.MerchantId,
			TerminalId:      info.TerminalId,
			ReferenceNumber: info.ReferenceNumber,
			RefundCounter:   info.RefundCounter,
		}
		if info.AvailableAmount != nil {
			modelResp.AdditionalInfo.AvailableAmount = &model.Amount{
				Value:    info.AvailableAmount.Value,
				Currency: info.AvailableAmount.Currency,
			}
		}
	}

	return modelResp
}
