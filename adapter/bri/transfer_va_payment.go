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

type briVirtualAccountPaymentBillDetail struct {
	BillNo          string                 `json:"billNo"`
	BillDescription *briLocalizedText      `json:"billDescription"`
	BillSubCompany  string                 `json:"billSubCompany"`
	BillAmount      *model.Amount          `json:"billAmount"`
	AdditionalInfo  map[string]interface{} `json:"additionalInfo"`
	BillReferenceNo string                 `json:"billReferenceNo"`
}

type briVirtualAccountPaymentRequest struct {
	PartnerServiceId        string                               `json:"partnerServiceId"`
	CustomerNo              string                               `json:"customerNo"`
	VirtualAccountNo        string                               `json:"virtualAccountNo"`
	VirtualAccountName      string                               `json:"virtualAccountName"`
	PaymentRequestId        string                               `json:"paymentRequestId"`
	ChannelCode             int                                  `json:"channelCode"`
	HashedSourceAccountNo   string                               `json:"hashedSourceAccountNo,omitempty"`
	SourceBankCode          string                               `json:"sourceBankCode,omitempty"`
	PaidAmount              *model.Amount                        `json:"paidAmount"`
	CumulativePaymentAmount *model.Amount                        `json:"cumulativePaymentAmount,omitempty"`
	PaidBills               string                               `json:"paidBills,omitempty"`
	TotalAmount             *model.Amount                        `json:"totalAmount,omitempty"`
	TrxDateTime             string                               `json:"trxDateTime,omitempty"`
	ReferenceNo             string                               `json:"referenceNo"`
	FlagAdvise              string                               `json:"flagAdvise"`
	SubCompany              string                               `json:"subCompany"`
	BillDetails             []briVirtualAccountPaymentBillDetail `json:"billDetails"`
	AdditionalInfo          map[string]interface{}               `json:"additionalInfo"`
}

type briVirtualAccountPaymentBillDetailResponse struct {
	BillerReferenceId string                 `json:"billerReferenceId"`
	BillNo            string                 `json:"billNo"`
	BillDescription   *briLocalizedText      `json:"billDescription"`
	BillSubCompany    string                 `json:"billSubCompany"`
	BillAmount        *model.Amount          `json:"billAmount"`
	AdditionalInfo    map[string]interface{} `json:"additionalInfo"`
	Status            string                 `json:"status"`
	Reason            *briLocalizedText      `json:"reason"`
}

type briVirtualAccountPaymentData struct {
	PaymentFlagReason  *briLocalizedText                            `json:"paymentFlagReason"`
	PartnerServiceId   string                                       `json:"partnerServiceId"`
	CustomerNo         string                                       `json:"customerNo"`
	VirtualAccountNo   string                                       `json:"virtualAccountNo"`
	VirtualAccountName string                                       `json:"virtualAccountName"`
	PaymentRequestId   string                                       `json:"paymentRequestId"`
	PaidAmount         *model.Amount                                `json:"paidAmount"`
	TotalAmount        *model.Amount                                `json:"totalAmount"`
	TrxDateTime        string                                       `json:"trxDateTime"`
	ReferenceNo        string                                       `json:"referenceNo"`
	PaymentFlagStatus  string                                       `json:"paymentFlagStatus"`
	BillDetails        []briVirtualAccountPaymentBillDetailResponse `json:"billDetails"`
}

type briVirtualAccountPaymentResponse struct {
	ResponseCode       string                           `json:"responseCode"`
	ResponseMessage    string                           `json:"responseMessage"`
	VirtualAccountData *briVirtualAccountPaymentData    `json:"virtualAccountData"`
	AdditionalInfo     map[string]briAdditionalInfoItem `json:"additionalInfo"`
}

func (a *BRIAdapter) GetVirtualAccountPayment(ctx context.Context, accessToken string, request *model.VirtualAccountPaymentRequest) (*model.VirtualAccountPaymentResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBRI.Name(), "GetVirtualAccountPayment")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointVirtualAccountPayment)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	billDetails := make([]briVirtualAccountPaymentBillDetail, 0, len(request.BillDetails))
	for _, b := range request.BillDetails {
		detail := briVirtualAccountPaymentBillDetail{
			BillNo:          b.BillNo,
			BillSubCompany:  b.BillSubCompany,
			AdditionalInfo:  b.AdditionalInfo,
			BillReferenceNo: b.BillReferenceNo,
		}
		if b.BillAmount != nil {
			detail.BillAmount = &model.Amount{
				Value:    b.BillAmount.Value,
				Currency: b.BillAmount.Currency,
			}
		}
		if b.BillDescription != nil {
			detail.BillDescription = &briLocalizedText{
				English:   b.BillDescription.English,
				Indonesia: b.BillDescription.Indonesia,
			}
		}
		billDetails = append(billDetails, detail)
	}

	requestBody := briVirtualAccountPaymentRequest{
		PartnerServiceId:   request.PartnerServiceId,
		CustomerNo:         request.CustomerNo,
		VirtualAccountNo:   request.VirtualAccountNo,
		VirtualAccountName: request.VirtualAccountName,
		PaymentRequestId:   request.PaymentRequestId,
		ChannelCode:        request.ChannelCode,
		ReferenceNo:        request.ReferenceNo,
		FlagAdvise:         request.FlagAdvise,
		SubCompany:         request.SubCompany,
		BillDetails:        billDetails,
		AdditionalInfo:     request.AdditionalInfo,
	}

	if request.HashedSourceAccountNo != "" {
		requestBody.HashedSourceAccountNo = request.HashedSourceAccountNo
	}
	if request.SourceBankCode != "" {
		requestBody.SourceBankCode = request.SourceBankCode
	}
	if request.PaidAmount != nil {
		requestBody.PaidAmount = &model.Amount{
			Value:    request.PaidAmount.Value,
			Currency: request.PaidAmount.Currency,
		}
	}
	if request.CumulativePaymentAmount != nil {
		requestBody.CumulativePaymentAmount = &model.Amount{
			Value:    request.CumulativePaymentAmount.Value,
			Currency: request.CumulativePaymentAmount.Currency,
		}
	}
	if request.TotalAmount != nil {
		requestBody.TotalAmount = &model.Amount{
			Value:    request.TotalAmount.Value,
			Currency: request.TotalAmount.Currency,
		}
	}
	if request.PaidBills != "" {
		requestBody.PaidBills = request.PaidBills
	}
	if !request.TrxDateTime.IsZero() {
		requestBody.TrxDateTime = request.TrxDateTime.Format(time.RFC3339)
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalVirtualAccountPaymentRequest,
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
			Operation: model.OperationVirtualAccountPaymentRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadVirtualAccountPaymentResponse,
			Err:       err,
		}
	}

	var resp briVirtualAccountPaymentResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalVirtualAccountPaymentResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessVirtualAccountPayment {
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

	return briVirtualAccountPaymentResponseToModel(resp), nil
}

func briVirtualAccountPaymentResponseToModel(briResp briVirtualAccountPaymentResponse) *model.VirtualAccountPaymentResponse {
	modelResp := &model.VirtualAccountPaymentResponse{
		ResponseCode:    briResp.ResponseCode,
		ResponseMessage: briResp.ResponseMessage,
		Raw:             briResp,
	}

	if len(briResp.AdditionalInfo) > 0 {
		modelResp.AdditionalInfo = make(map[string]model.AdditionalInfoItem, len(briResp.AdditionalInfo))
		for key, item := range briResp.AdditionalInfo {
			modelResp.AdditionalInfo[key] = model.AdditionalInfoItem{
				Label: parseBRILocalizedText(item.Label),
				Value: parseBRILocalizedText(item.Value),
			}
		}
	}

	if briResp.VirtualAccountData == nil {
		return modelResp
	}

	data := briResp.VirtualAccountData
	modelResp.VirtualAccountData = &model.VirtualAccountPaymentData{
		PaymentFlagReason:  parseBRILocalizedText(data.PaymentFlagReason),
		PartnerServiceId:   data.PartnerServiceId,
		CustomerNo:         data.CustomerNo,
		VirtualAccountNo:   data.VirtualAccountNo,
		VirtualAccountName: data.VirtualAccountName,
		PaymentRequestId:   data.PaymentRequestId,
		TrxDateTime:        data.TrxDateTime,
		ReferenceNo:        data.ReferenceNo,
		PaymentFlagStatus:  data.PaymentFlagStatus,
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
			BillerReferenceId: b.BillerReferenceId,
			BillNo:            b.BillNo,
			BillSubCompany:    b.BillSubCompany,
			AdditionalInfo:    b.AdditionalInfo,
			Status:            b.Status,
			BillDescription:   parseBRILocalizedText(b.BillDescription),
			Reason:            parseBRILocalizedText(b.Reason),
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
