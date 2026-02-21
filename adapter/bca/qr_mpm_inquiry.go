package bca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"snap-banking-lib/internal/utils"
	"snap-banking-lib/model"
)

type bcaQRMPMInquiryAdditionalInfo struct {
	TerminalId          string `json:"terminalId"`
	PartnerMerchantType string `json:"partnerMerchantType"`
}

type bcaQRMPMInquiryRequest struct {
	OriginalPartnerReferenceNo string                         `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo        string                         `json:"originalReferenceNo"`
	ServiceCode                string                         `json:"serviceCode"`
	MerchantId                 string                         `json:"merchantId"`
	SubMerchantId              string                         `json:"subMerchantId"`
	AdditionalInfo             *bcaQRMPMInquiryAdditionalInfo `json:"additionalInfo"`
}

type bcaQRMPMInquiryMerchantInfo struct {
	MerchantId         string  `json:"merchantId"`
	MerchantPan        string  `json:"merchantPan"`
	Name               string  `json:"name"`
	City               string  `json:"city"`
	PostalCode         string  `json:"postalCode"`
	Country            string  `json:"country"`
	Email              *string `json:"email"`
	PaymentChannelName string  `json:"paymentChannelName"`
}

type bcaQRMPMInquiryResponseAdditionalInfo struct {
	ReferenceNumber       string                       `json:"referenceNumber"`
	ApprovalCode          *string                      `json:"approvalCode"`
	PayerPhoneNumber      *string                      `json:"payerPhoneNumber"`
	BatchNumber           *string                      `json:"batchNumber"`
	ConvenienceFee        *string                      `json:"convenienceFee"`
	IssuerReferenceNumber *string                      `json:"issuerReferenceNumber"`
	PayerName             *string                      `json:"payerName"`
	CustomerPan           *string                      `json:"customerPan"`
	IssuerName            *string                      `json:"issuerName"`
	AcquirerName          *string                      `json:"acquirerName"`
	MerchantInfo          *bcaQRMPMInquiryMerchantInfo `json:"merchantInfo"`
}

type bcaQRMPMInquiryResponse struct {
	ResponseCode               string                                 `json:"responseCode"`
	ResponseMessage            string                                 `json:"responseMessage"`
	OriginalPartnerReferenceNo string                                 `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo        string                                 `json:"originalReferenceNo"`
	OriginalExternalId         string                                 `json:"originalExternalId"`
	ServiceCode                string                                 `json:"serviceCode"`
	LatestTransactionStatus    string                                 `json:"latestTransactionStatus"`
	TransactionStatusDesc      string                                 `json:"transactionStatusDesc"`
	PaidTime                   *string                                `json:"paidTime"`
	Amount                     *model.Amount                          `json:"amount"`
	FeeAmount                  *model.Amount                          `json:"feeAmount"`
	TerminalId                 *string                                `json:"terminalId"`
	AdditionalInfo             *bcaQRMPMInquiryResponseAdditionalInfo `json:"additionalInfo"`
}

func (a *BCAAdapter) GetQRMPMInquiry(ctx context.Context, accessToken string, request *model.QRMPMInquiryRequest) (*model.QRMPMInquiryResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "GetQRMPMInquiry")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointQRMPMInquiry)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := bcaQRMPMInquiryRequest{
		OriginalPartnerReferenceNo: request.OriginalPartnerReferenceNo,
		OriginalReferenceNo:        request.OriginalReferenceNo,
		ServiceCode:                request.ServiceCode,
		MerchantId:                 request.MerchantId,
		SubMerchantId:              request.SubMerchantId,
	}

	if request.AdditionalInfo != nil {
		requestBody.AdditionalInfo = &bcaQRMPMInquiryAdditionalInfo{
			TerminalId:          request.AdditionalInfo.TerminalId,
			PartnerMerchantType: request.AdditionalInfo.PartnerMerchantType,
		}
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalQRMPMInquiryRequest,
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
			Operation: model.OperationQRMPMInquiryRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadQRMPMInquiryResponse,
			Err:       err,
		}
	}

	var resp bcaQRMPMInquiryResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalQRMPMInquiryResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessQRMPMInquiry {
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

	return bcaQRMPMInquiryResponseToModel(resp), nil
}

func bcaQRMPMInquiryResponseToModel(bcaResp bcaQRMPMInquiryResponse) *model.QRMPMInquiryResponse {
	modelResp := &model.QRMPMInquiryResponse{
		ResponseCode:               bcaResp.ResponseCode,
		ResponseMessage:            bcaResp.ResponseMessage,
		OriginalPartnerReferenceNo: bcaResp.OriginalPartnerReferenceNo,
		OriginalReferenceNo:        bcaResp.OriginalReferenceNo,
		OriginalExternalId:         bcaResp.OriginalExternalId,
		ServiceCode:                bcaResp.ServiceCode,
		LatestTransactionStatus:    bcaResp.LatestTransactionStatus,
		TransactionStatusDesc:      bcaResp.TransactionStatusDesc,
		PaidTime:                   bcaResp.PaidTime,
		TerminalId:                 bcaResp.TerminalId,
		Raw:                        bcaResp,
	}

	if bcaResp.Amount != nil {
		modelResp.Amount = &model.Amount{
			Value:    bcaResp.Amount.Value,
			Currency: bcaResp.Amount.Currency,
		}
	}

	if bcaResp.FeeAmount != nil {
		modelResp.FeeAmount = &model.Amount{
			Value:    bcaResp.FeeAmount.Value,
			Currency: bcaResp.FeeAmount.Currency,
		}
	}

	if bcaResp.AdditionalInfo != nil {
		info := bcaResp.AdditionalInfo
		modelResp.AdditionalInfo = &model.QRMPMInquiryResponseAdditionalInfo{
			ReferenceNumber:       info.ReferenceNumber,
			ApprovalCode:          info.ApprovalCode,
			PayerPhoneNumber:      info.PayerPhoneNumber,
			BatchNumber:           info.BatchNumber,
			ConvenienceFee:        info.ConvenienceFee,
			IssuerReferenceNumber: info.IssuerReferenceNumber,
			PayerName:             info.PayerName,
			CustomerPan:           info.CustomerPan,
			IssuerName:            info.IssuerName,
			AcquirerName:          info.AcquirerName,
		}

		if info.MerchantInfo != nil {
			modelResp.AdditionalInfo.MerchantInfo = &model.QRMPMInquiryMerchantInfo{
				MerchantId:         info.MerchantInfo.MerchantId,
				MerchantPan:        info.MerchantInfo.MerchantPan,
				Name:               info.MerchantInfo.Name,
				City:               info.MerchantInfo.City,
				PostalCode:         info.MerchantInfo.PostalCode,
				Country:            info.MerchantInfo.Country,
				Email:              info.MerchantInfo.Email,
				PaymentChannelName: info.MerchantInfo.PaymentChannelName,
			}
		}
	}

	return modelResp
}
