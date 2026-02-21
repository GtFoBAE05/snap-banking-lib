package bca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"snap-banking-lib/internal/utils"
	"snap-banking-lib/model"
)

type bcaVirtualAccountInquiryStatusRequest struct {
	PartnerServiceId string                 `json:"partnerServiceId"`
	CustomerNo       string                 `json:"customerNo"`
	VirtualAccountNo string                 `json:"virtualAccountNo"`
	PaymentRequestId string                 `json:"paymentRequestId"`
	AdditionalInfo   map[string]interface{} `json:"additionalInfo"`
}

type bcaVirtualAccountInquiryStatusBillDetail struct {
	BillNo          string                 `json:"billNo"`
	BillDescription *bcaLocalizedText      `json:"billDescription"`
	BillSubCompany  string                 `json:"billSubCompany"`
	BillAmount      *model.Amount          `json:"billAmount"`
	AdditionalInfo  map[string]interface{} `json:"additionalInfo"`
	BillReferenceNo string                 `json:"billReferenceNo"`
	Status          string                 `json:"status"`
	Reason          *bcaLocalizedText      `json:"reason"`
}

type bcaVirtualAccountInquiryStatusData struct {
	PaymentFlagStatus string                                     `json:"paymentFlagStatus"`
	PaymentFlagReason *bcaLocalizedText                          `json:"paymentFlagReason"`
	PartnerServiceId  string                                     `json:"partnerServiceId"`
	CustomerNo        string                                     `json:"customerNo"`
	VirtualAccountNo  string                                     `json:"virtualAccountNo"`
	InquiryRequestId  string                                     `json:"inquiryRequestId"`
	PaymentRequestId  string                                     `json:"paymentRequestId"`
	PaidAmount        *model.Amount                              `json:"paidAmount"`
	TotalAmount       *model.Amount                              `json:"totalAmount"`
	TransactionDate   string                                     `json:"transactionDate"`
	ReferenceNo       string                                     `json:"referenceNo"`
	BillDetails       []bcaVirtualAccountInquiryStatusBillDetail `json:"billDetails"`
}

type bcaVirtualAccountInquiryStatusResponse struct {
	ResponseCode       string                              `json:"responseCode"`
	ResponseMessage    string                              `json:"responseMessage"`
	VirtualAccountData *bcaVirtualAccountInquiryStatusData `json:"virtualAccountData"`
}

func (a *BCAAdapter) GetVirtualAccountInquiryStatus(ctx context.Context, accessToken string, request *model.VirtualAccountInquiryStatusRequest) (*model.VirtualAccountInquiryStatusResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "GetVirtualAccountInquiryStatus")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointVirtualAccountInquiryStatus)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := bcaVirtualAccountInquiryStatusRequest{
		PartnerServiceId: request.PartnerServiceId,
		CustomerNo:       request.CustomerNo,
		VirtualAccountNo: request.VirtualAccountNo,
		PaymentRequestId: request.PaymentRequestId,
		AdditionalInfo:   request.AdditionalInfo,
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalVirtualAccountInquiryStatusRequest,
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
			Operation: model.OperationVirtualAccountInquiryStatusRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadVirtualAccountInquiryStatusResponse,
			Err:       err,
		}
	}

	var resp bcaVirtualAccountInquiryStatusResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalVirtualAccountInquiryStatusResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessVirtualAccountInquiryStatus {
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

	return bcaVirtualAccountInquiryStatusResponseToModel(resp), nil
}

func bcaVirtualAccountInquiryStatusResponseToModel(bcaResp bcaVirtualAccountInquiryStatusResponse) *model.VirtualAccountInquiryStatusResponse {
	modelResp := &model.VirtualAccountInquiryStatusResponse{
		ResponseCode:    bcaResp.ResponseCode,
		ResponseMessage: bcaResp.ResponseMessage,
		Raw:             bcaResp,
	}

	if bcaResp.VirtualAccountData == nil {
		return modelResp
	}

	data := bcaResp.VirtualAccountData
	modelResp.VirtualAccountData = &model.VirtualAccountInquiryStatusData{
		PaymentFlagStatus: data.PaymentFlagStatus,
		PaymentFlagReason: parseLocalizedText(data.PaymentFlagReason),
		PartnerServiceId:  data.PartnerServiceId,
		CustomerNo:        data.CustomerNo,
		VirtualAccountNo:  data.VirtualAccountNo,
		InquiryRequestId:  data.InquiryRequestId,
		PaymentRequestId:  data.PaymentRequestId,
		TransactionDate:   data.TransactionDate,
		ReferenceNo:       data.ReferenceNo,
	}

	if data.PaidAmount != nil {
		modelResp.VirtualAccountData.PaidAmount = &model.Amount{
			Value:    data.PaidAmount.Value,
			Currency: data.PaidAmount.Currency,
		}
	}

	if data.TotalAmount != nil {
		modelResp.VirtualAccountData.TotalAmount = &model.Amount{
			Value:    data.TotalAmount.Value,
			Currency: data.TotalAmount.Currency,
		}
	}

	for _, b := range data.BillDetails {
		detail := model.BillDetail{
			BillNo:          b.BillNo,
			BillSubCompany:  b.BillSubCompany,
			BillReferenceNo: b.BillReferenceNo,
			AdditionalInfo:  b.AdditionalInfo,
			Status:          b.Status,
			BillDescription: parseLocalizedText(b.BillDescription),
			Reason:          parseLocalizedText(b.Reason),
		}
		if b.BillAmount != nil {
			detail.BillAmount = &model.Amount{
				Value:    b.BillAmount.Value,
				Currency: b.BillAmount.Currency,
			}
		}
		modelResp.VirtualAccountData.BillDetails = append(modelResp.VirtualAccountData.BillDetails, detail)
	}

	return modelResp
}
