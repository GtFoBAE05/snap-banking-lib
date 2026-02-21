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

type bcaQRMPMGenerateAdditionalInfo struct {
	ConvenienceFee       string `json:"convenienceFee"`
	PartnerMerchantType  string `json:"partnerMerchantType"`
	TerminalLocationName string `json:"terminalLocationName"`
	QrOption             string `json:"qrOption"`
}

type bcaQRMPMGenerateRequest struct {
	PartnerReferenceNo string                          `json:"partnerReferenceNo"`
	Amount             *model.Amount                   `json:"amount"`
	MerchantId         string                          `json:"merchantId"`
	SubMerchantId      string                          `json:"subMerchantId"`
	TerminalId         string                          `json:"terminalId"`
	ValidityPeriod     string                          `json:"validityPeriod"`
	AdditionalInfo     *bcaQRMPMGenerateAdditionalInfo `json:"additionalInfo"`
}

type bcaQRMPMGenerateResponse struct {
	ResponseCode       string      `json:"responseCode"`
	ResponseMessage    string      `json:"responseMessage"`
	ReferenceNo        string      `json:"referenceNo"`
	PartnerReferenceNo string      `json:"partnerReferenceNo"`
	QrContent          string      `json:"qrContent"`
	QrUrl              *string     `json:"qrUrl"`
	QrImage            string      `json:"qrImage"`
	RedirectUrl        *string     `json:"redirectUrl"`
	MerchantName       string      `json:"merchantName"`
	StoreId            *string     `json:"storeId"`
	TerminalId         string      `json:"terminalId"`
	AdditionalInfo     interface{} `json:"additionalInfo"`
}

func (a *BCAAdapter) GenerateQRMPM(ctx context.Context, accessToken string, request *model.QRMPMGenerateRequest) (*model.QRMPMGenerateResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "GenerateQRMPM")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointQRMPMGenerate)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := bcaQRMPMGenerateRequest{
		PartnerReferenceNo: request.PartnerReferenceNo,
		MerchantId:         request.MerchantId,
		SubMerchantId:      request.SubMerchantId,
		TerminalId:         request.TerminalId,
		ValidityPeriod:     request.ValidityPeriod.Format(time.RFC3339),
	}

	if request.Amount != nil {
		requestBody.Amount = &model.Amount{
			Value:    request.Amount.Value,
			Currency: request.Amount.Currency,
		}
	}

	if request.AdditionalInfo != nil {
		requestBody.AdditionalInfo = &bcaQRMPMGenerateAdditionalInfo{
			ConvenienceFee:       request.AdditionalInfo.ConvenienceFee,
			PartnerMerchantType:  request.AdditionalInfo.PartnerMerchantType,
			TerminalLocationName: request.AdditionalInfo.TerminalLocationName,
			QrOption:             request.AdditionalInfo.QrOption,
		}
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalQRMPMGenerateRequest,
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
			Operation: model.OperationQRMPMGenerateRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadQRMPMGenerateResponse,
			Err:       err,
		}
	}

	var resp bcaQRMPMGenerateResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalQRMPMGenerateResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessQRMPMGenerate {
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

	return bcaQRMPMGenerateResponseToModel(resp), nil
}

func bcaQRMPMGenerateResponseToModel(bcaResp bcaQRMPMGenerateResponse) *model.QRMPMGenerateResponse {
	return &model.QRMPMGenerateResponse{
		ResponseCode:       bcaResp.ResponseCode,
		ResponseMessage:    bcaResp.ResponseMessage,
		ReferenceNo:        bcaResp.ReferenceNo,
		PartnerReferenceNo: bcaResp.PartnerReferenceNo,
		QrContent:          bcaResp.QrContent,
		QrUrl:              bcaResp.QrUrl,
		QrImage:            bcaResp.QrImage,
		RedirectUrl:        bcaResp.RedirectUrl,
		MerchantName:       bcaResp.MerchantName,
		StoreId:            bcaResp.StoreId,
		TerminalId:         bcaResp.TerminalId,
		AdditionalInfo:     bcaResp.AdditionalInfo,
		Raw:                bcaResp,
	}
}
