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

type bcaVirtualAccountIntrabankPaymentRequest struct {
	VirtualAccountNo    string        `json:"virtualAccountNo"`
	VirtualAccountEmail string        `json:"virtualAccountEmail"`
	SourceAccountNo     string        `json:"sourceAccountNo"`
	PartnerReferenceNo  string        `json:"partnerReferenceNo"`
	PaidAmount          *model.Amount `json:"paidAmount"`
	TrxDateTime         string        `json:"trxDateTime"`
}

type bcaVirtualAccountIntrabankPaymentData struct {
	VirtualAccountNo    string                   `json:"virtualAccountNo"`
	VirtualAccountName  string                   `json:"virtualAccountName"`
	VirtualAccountEmail string                   `json:"virtualAccountEmail"`
	SourceAccountNo     string                   `json:"sourceAccountNo"`
	PartnerReferenceNo  string                   `json:"partnerReferenceNo"`
	ReferenceNo         string                   `json:"referenceNo"`
	PaidAmount          *model.Amount            `json:"paidAmount"`
	TotalAmount         *model.Amount            `json:"totalAmount"`
	TrxDateTime         string                   `json:"trxDateTime"`
	BillDetails         []bcaIntrabankBillDetail `json:"billDetails"`
	FreeTexts           []bcaLocalizedText       `json:"freeTexts"`
	FeeAmount           *model.Amount            `json:"feeAmount"`
	ProductName         string                   `json:"productName"`
}

type bcaVirtualAccountIntrabankPaymentResponse struct {
	ResponseCode       string                                 `json:"responseCode"`
	ResponseMessage    string                                 `json:"responseMessage"`
	VirtualAccountData *bcaVirtualAccountIntrabankPaymentData `json:"virtualAccountData"`
}

func (a *BCAAdapter) GetVirtualAccountIntrabankPayment(ctx context.Context, accessToken string, request *model.VirtualAccountIntrabankPaymentRequest) (*model.VirtualAccountIntrabankPaymentResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "GetVirtualAccountIntrabankPayment")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointVirtualAccountIntrabankPayment)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := bcaVirtualAccountIntrabankPaymentRequest{
		VirtualAccountNo:    request.VirtualAccountNo,
		VirtualAccountEmail: request.VirtualAccountEmail,
		SourceAccountNo:     request.SourceAccountNo,
		PartnerReferenceNo:  request.PartnerReferenceNo,
		TrxDateTime:         request.TrxDateTime.Format(time.RFC3339),
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

	var resp bcaVirtualAccountIntrabankPaymentResponse
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

	return bcaVirtualAccountIntrabankPaymentResponseToModel(resp), nil
}

func bcaVirtualAccountIntrabankPaymentResponseToModel(bcaResp bcaVirtualAccountIntrabankPaymentResponse) *model.VirtualAccountIntrabankPaymentResponse {
	modelResp := &model.VirtualAccountIntrabankPaymentResponse{
		ResponseCode:    bcaResp.ResponseCode,
		ResponseMessage: bcaResp.ResponseMessage,
		Raw:             bcaResp,
	}

	if bcaResp.VirtualAccountData == nil {
		return modelResp
	}

	data := bcaResp.VirtualAccountData
	modelResp.VirtualAccountData = &model.VirtualAccountIntrabankPaymentData{
		VirtualAccountNo:    data.VirtualAccountNo,
		VirtualAccountName:  data.VirtualAccountName,
		VirtualAccountEmail: data.VirtualAccountEmail,
		SourceAccountNo:     data.SourceAccountNo,
		PartnerReferenceNo:  data.PartnerReferenceNo,
		ReferenceNo:         data.ReferenceNo,
		TrxDateTime:         data.TrxDateTime,
		ProductName:         data.ProductName,
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

	if data.FeeAmount != nil {
		modelResp.VirtualAccountData.FeeAmount = &model.Amount{
			Value:    data.FeeAmount.Value,
			Currency: data.FeeAmount.Currency,
		}
	}

	for _, b := range data.BillDetails {
		detail := model.IntrabankBillDetail{
			BillDescription: parseLocalizedText(b.BillDescription),
		}
		if b.BillAmount != nil {
			detail.BillAmount = &model.Amount{
				Value:    b.BillAmount.Value,
				Currency: b.BillAmount.Currency,
			}
		}
		modelResp.VirtualAccountData.BillDetails = append(modelResp.VirtualAccountData.BillDetails, detail)
	}

	for _, f := range data.FreeTexts {
		modelResp.VirtualAccountData.FreeTexts = append(modelResp.VirtualAccountData.FreeTexts, model.LocalizedText{
			English:   f.English,
			Indonesia: f.Indonesia,
		})
	}

	return modelResp
}
