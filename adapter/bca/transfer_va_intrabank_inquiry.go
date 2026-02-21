package bca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"snap-banking-lib/internal/utils"
	"snap-banking-lib/model"
)

type bcaVirtualAccountIntrabankInquiryRequest struct {
	VirtualAccountNo string        `json:"virtualAccountNo"`
	Amount           *model.Amount `json:"amount"`
}

type bcaIntrabankBillDetail struct {
	BillDescription *bcaLocalizedText `json:"billDescription"`
	BillAmount      *model.Amount     `json:"billAmount"`
}

type bcaVirtualAccountIntrabankInquiryData struct {
	VirtualAccountNo      string                   `json:"virtualAccountNo"`
	VirtualAccountName    string                   `json:"virtualAccountName"`
	TotalAmount           *model.Amount            `json:"totalAmount"`
	BillDetails           []bcaIntrabankBillDetail `json:"billDetails"`
	FreeTexts             []bcaLocalizedText       `json:"freeTexts"`
	VirtualAccountTrxType string                   `json:"virtualAccountTrxType"`
	FeeAmount             *model.Amount            `json:"feeAmount"`
	ProductName           string                   `json:"productName"`
}

type bcaVirtualAccountIntrabankInquiryResponse struct {
	ResponseCode       string                                 `json:"responseCode"`
	ResponseMessage    string                                 `json:"responseMessage"`
	VirtualAccountData *bcaVirtualAccountIntrabankInquiryData `json:"virtualAccountData"`
}

func (a *BCAAdapter) GetVirtualAccountIntrabankInquiry(ctx context.Context, accessToken string, request *model.VirtualAccountIntrabankInquiryRequest) (*model.VirtualAccountIntrabankInquiryResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "GetVirtualAccountIntrabankInquiry")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointVirtualAccountIntrabankInquiry)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := bcaVirtualAccountIntrabankInquiryRequest{
		VirtualAccountNo: request.VirtualAccountNo,
	}

	if request.Amount != nil {
		requestBody.Amount = &model.Amount{
			Value:    request.Amount.Value,
			Currency: request.Amount.Currency,
		}
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalVirtualAccountIntrabankInquiryRequest,
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
			Operation: model.OperationVirtualAccountIntrabankInquiryRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadVirtualAccountIntrabankInquiryResponse,
			Err:       err,
		}
	}

	var resp bcaVirtualAccountIntrabankInquiryResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalVirtualAccountIntrabankInquiryResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessVirtualAccountIntrabankInquiry {
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

	return bcaVirtualAccountIntrabankInquiryResponseToModel(resp), nil
}

func bcaVirtualAccountIntrabankInquiryResponseToModel(bcaResp bcaVirtualAccountIntrabankInquiryResponse) *model.VirtualAccountIntrabankInquiryResponse {
	modelResp := &model.VirtualAccountIntrabankInquiryResponse{
		ResponseCode:    bcaResp.ResponseCode,
		ResponseMessage: bcaResp.ResponseMessage,
		Raw:             bcaResp,
	}

	if bcaResp.VirtualAccountData == nil {
		return modelResp
	}

	data := bcaResp.VirtualAccountData
	modelResp.VirtualAccountData = &model.VirtualAccountIntrabankInquiryData{
		VirtualAccountNo:      data.VirtualAccountNo,
		VirtualAccountName:    data.VirtualAccountName,
		VirtualAccountTrxType: data.VirtualAccountTrxType,
		ProductName:           data.ProductName,
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
