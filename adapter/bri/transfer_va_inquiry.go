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

type briVirtualAccountInquiryRequest struct {
	PartnerServiceId string                 `json:"partnerServiceId"`
	CustomerNo       string                 `json:"customerNo"`
	VirtualAccountNo string                 `json:"virtualAccountNo"`
	TrxDateInit      string                 `json:"trxDateInit"`
	ChannelCode      int                    `json:"channelCode"`
	AdditionalInfo   map[string]interface{} `json:"additionalInfo"`
	InquiryRequestId string                 `json:"inquiryRequestId"`
}

type briLocalizedText struct {
	English   string `json:"english"`
	Indonesia string `json:"indonesia"`
}

type briBillDetail struct {
	BillNo          string            `json:"billNo"`
	BillDescription *briLocalizedText `json:"billDescription"`
	BillSubCompany  string            `json:"billSubCompany"`
	BillAmount      *model.Amount     `json:"billAmount"`
}

type briAdditionalInfoItem struct {
	Label *briLocalizedText `json:"label"`
	Value *briLocalizedText `json:"value"`
}

type briVirtualAccountData struct {
	InquiryStatus         string                           `json:"inquiryStatus"`
	InquiryReason         *briLocalizedText                `json:"inquiryReason"`
	PartnerServiceId      string                           `json:"partnerServiceId"`
	CustomerNo            string                           `json:"customerNo"`
	VirtualAccountNo      string                           `json:"virtualAccountNo"`
	VirtualAccountName    string                           `json:"virtualAccountName"`
	InquiryRequestId      string                           `json:"inquiryRequestId"`
	TotalAmount           *model.Amount                    `json:"totalAmount"`
	SubCompany            string                           `json:"subCompany"`
	BillDetails           []briBillDetail                  `json:"billDetails"`
	FreeTexts             []briLocalizedText               `json:"freeTexts"`
	VirtualAccountTrxType string                           `json:"virtualAccountTrxType"`
	FeeAmount             *model.Amount                    `json:"feeAmount"`
	AdditionalInfo        map[string]briAdditionalInfoItem `json:"additionalInfo"`
}

type briVirtualAccountInquiryResponse struct {
	ResponseCode       string                 `json:"responseCode"`
	ResponseMessage    string                 `json:"responseMessage"`
	VirtualAccountData *briVirtualAccountData `json:"virtualAccountData"`
}

func (a *BRIAdapter) GetVirtualAccountInquiry(ctx context.Context, accessToken string, request *model.VirtualAccountInquiryRequest) (*model.VirtualAccountInquiryResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBRI.Name(), "GetVirtualAccountInquiry")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointVirtualAccountInquiry)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := briVirtualAccountInquiryRequest{
		PartnerServiceId: request.PartnerServiceId,
		CustomerNo:       request.CustomerNo,
		VirtualAccountNo: request.VirtualAccountNo,
		TrxDateInit:      request.TrxDateInit.Format(time.RFC3339),
		ChannelCode:      request.ChannelCode,
		AdditionalInfo:   request.AdditionalInfo,
		InquiryRequestId: request.InquiryRequestId,
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalVirtualAccountInquiryRequest,
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
			Operation: model.OperationVirtualAccountInquiryRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadVirtualAccountInquiryResponse,
			Err:       err,
		}
	}

	var resp briVirtualAccountInquiryResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalVirtualAccountInquiryResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessVirtualAccountInquiry {
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

	return briVirtualAccountInquiryResponseToModel(resp), nil
}

func briVirtualAccountInquiryResponseToModel(briResp briVirtualAccountInquiryResponse) *model.VirtualAccountInquiryResponse {
	modelResp := &model.VirtualAccountInquiryResponse{
		ResponseCode:    briResp.ResponseCode,
		ResponseMessage: briResp.ResponseMessage,
		Raw:             briResp,
	}

	if briResp.VirtualAccountData == nil {
		return modelResp
	}

	data := briResp.VirtualAccountData
	modelResp.VirtualAccountData = &model.VirtualAccountData{
		InquiryStatus:         data.InquiryStatus,
		InquiryReason:         parseBRILocalizedText(data.InquiryReason),
		PartnerServiceId:      data.PartnerServiceId,
		CustomerNo:            data.CustomerNo,
		VirtualAccountNo:      data.VirtualAccountNo,
		VirtualAccountName:    data.VirtualAccountName,
		InquiryRequestId:      data.InquiryRequestId,
		SubCompany:            data.SubCompany,
		VirtualAccountTrxType: data.VirtualAccountTrxType,
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
		modelResp.VirtualAccountData.BillDetails = append(modelResp.VirtualAccountData.BillDetails, model.BillDetail{
			BillNo:          b.BillNo,
			BillDescription: parseBRILocalizedText(b.BillDescription),
			BillSubCompany:  b.BillSubCompany,
			BillAmount: func() *model.Amount {
				if b.BillAmount == nil {
					return nil
				}
				return &model.Amount{
					Value:    b.BillAmount.Value,
					Currency: b.BillAmount.Currency,
				}
			}(),
		})
	}

	for _, f := range data.FreeTexts {
		modelResp.VirtualAccountData.FreeTexts = append(modelResp.VirtualAccountData.FreeTexts, model.LocalizedText{
			English:   f.English,
			Indonesia: f.Indonesia,
		})
	}

	if len(data.AdditionalInfo) > 0 {
		modelResp.VirtualAccountData.AdditionalInfo = make(map[string]model.AdditionalInfoItem, len(data.AdditionalInfo))
		for key, item := range data.AdditionalInfo {
			modelResp.VirtualAccountData.AdditionalInfo[key] = model.AdditionalInfoItem{
				Label: parseBRILocalizedText(item.Label),
				Value: parseBRILocalizedText(item.Value),
			}
		}
	}

	return modelResp
}

func parseBRILocalizedText(t *briLocalizedText) *model.LocalizedText {
	if t == nil {
		return nil
	}
	return &model.LocalizedText{
		English:   t.English,
		Indonesia: t.Indonesia,
	}
}
