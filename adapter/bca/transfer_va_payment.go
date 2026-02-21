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

type bcaVirtualAccountPaymentBillDetail struct {
	BillNo          string                 `json:"billNo"`
	BillDescription *bcaLocalizedText      `json:"billDescription"`
	BillSubCompany  string                 `json:"billSubCompany"`
	BillAmount      *model.Amount          `json:"billAmount"`
	AdditionalInfo  map[string]interface{} `json:"additionalInfo"`
	BillReferenceNo string                 `json:"billReferenceNo"`
}

type bcaVirtualAccountPaymentRequest struct {
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
	BillDetails             []bcaVirtualAccountPaymentBillDetail `json:"billDetails"`
	AdditionalInfo          map[string]interface{}               `json:"additionalInfo"`
}

type bcaVirtualAccountPaymentBillDetailResponse struct {
	BillerReferenceId string                 `json:"billerReferenceId"`
	BillNo            string                 `json:"billNo"`
	BillDescription   *bcaLocalizedText      `json:"billDescription"`
	BillSubCompany    string                 `json:"billSubCompany"`
	BillAmount        *model.Amount          `json:"billAmount"`
	AdditionalInfo    map[string]interface{} `json:"additionalInfo"`
	Status            string                 `json:"status"`
	Reason            *bcaLocalizedText      `json:"reason"`
}

type bcaVirtualAccountPaymentData struct {
	PaymentFlagReason  *bcaLocalizedText                            `json:"paymentFlagReason"`
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
	BillDetails        []bcaVirtualAccountPaymentBillDetailResponse `json:"billDetails"`
}

type bcaVirtualAccountPaymentResponse struct {
	ResponseCode       string                           `json:"responseCode"`
	ResponseMessage    string                           `json:"responseMessage"`
	VirtualAccountData *bcaVirtualAccountPaymentData    `json:"virtualAccountData"`
	AdditionalInfo     map[string]bcaAdditionalInfoItem `json:"additionalInfo"`
}

func (a *BCAAdapter) GetVirtualAccountPayment(ctx context.Context, accessToken string, request *model.VirtualAccountPaymentRequest) (*model.VirtualAccountPaymentResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "GetVirtualAccountPayment")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointVirtualAccountPayment)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	billDetails := make([]bcaVirtualAccountPaymentBillDetail, 0, len(request.BillDetails))
	for _, b := range request.BillDetails {
		detail := bcaVirtualAccountPaymentBillDetail{
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
			detail.BillDescription = &bcaLocalizedText{
				English:   b.BillDescription.English,
				Indonesia: b.BillDescription.Indonesia,
			}
		}
		billDetails = append(billDetails, detail)
	}

	requestBody := bcaVirtualAccountPaymentRequest{
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

	var resp bcaVirtualAccountPaymentResponse
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

	return bcaVirtualAccountPaymentResponseToModel(resp), nil
}

func bcaVirtualAccountPaymentResponseToModel(bcaResp bcaVirtualAccountPaymentResponse) *model.VirtualAccountPaymentResponse {
	modelResp := &model.VirtualAccountPaymentResponse{
		ResponseCode:    bcaResp.ResponseCode,
		ResponseMessage: bcaResp.ResponseMessage,
		Raw:             bcaResp,
	}

	if len(bcaResp.AdditionalInfo) > 0 {
		modelResp.AdditionalInfo = make(map[string]model.AdditionalInfoItem, len(bcaResp.AdditionalInfo))
		for key, item := range bcaResp.AdditionalInfo {
			modelResp.AdditionalInfo[key] = model.AdditionalInfoItem{
				Label: parseLocalizedText(item.Label),
				Value: parseLocalizedText(item.Value),
			}
		}
	}

	if bcaResp.VirtualAccountData == nil {
		return modelResp
	}

	data := bcaResp.VirtualAccountData
	modelResp.VirtualAccountData = &model.VirtualAccountPaymentData{
		PaymentFlagReason:  parseLocalizedText(data.PaymentFlagReason),
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
			BillDescription:   parseLocalizedText(b.BillDescription),
			Reason:            parseLocalizedText(b.Reason),
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
