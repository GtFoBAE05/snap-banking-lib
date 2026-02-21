package bri

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"snap-banking-lib/internal/utils"
	"snap-banking-lib/model"
)

type briBalanceInquiryRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	AccountNo          string `json:"accountNo"`
}

type briAccountInfo struct {
	Amount           *model.Amount `json:"amount"`
	FloatAmount      *model.Amount `json:"floatAmount"`
	HoldAmount       *model.Amount `json:"holdAmount"`
	AvailableBalance *model.Amount `json:"availableBalance"`
	LedgerBalance    *model.Amount `json:"ledgerBalance"`
	Status           string        `json:"status"`
}

type briBalanceInquiryAdditionalInfo struct {
	ProductCode string `json:"productCode"`
	AccountType string `json:"accountType"`
}

type briBalanceInquiryResponse struct {
	ResponseCode       string                           `json:"responseCode"`
	ResponseMessage    string                           `json:"responseMessage"`
	ReferenceNo        string                           `json:"referenceNo"`
	PartnerReferenceNo string                           `json:"partnerReferenceNo"`
	AccountNo          string                           `json:"accountNo"`
	Name               string                           `json:"name"`
	AccountInfos       []briAccountInfo                 `json:"accountInfos"`
	AdditionalInfo     *briBalanceInquiryAdditionalInfo `json:"additionalInfo"`
}

func (a *BRIAdapter) GetBalanceInquiry(ctx context.Context, accessToken string, request *model.BalanceInquiryRequest) (*model.BalanceInquiryResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBRI.Name(), "GetBalanceInquiry")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointBalanceInquiry)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := briBalanceInquiryRequest{
		PartnerReferenceNo: request.PartnerReferenceNo,
		AccountNo:          request.AccountNo,
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalBalanceInquiryRequest,
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
			Operation: model.OperationBalanceInquiryRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadBalanceInquiryResponse,
			Err:       err,
		}
	}

	var resp briBalanceInquiryResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalBalanceInquiryResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessBalanceInquiry {
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

	return briBalanceInquiryResponseToModel(resp), nil
}

func briBalanceInquiryResponseToModel(briResp briBalanceInquiryResponse) *model.BalanceInquiryResponse {
	modelResp := &model.BalanceInquiryResponse{
		ResponseCode:       briResp.ResponseCode,
		ResponseMessage:    briResp.ResponseMessage,
		ReferenceNo:        briResp.ReferenceNo,
		PartnerReferenceNo: briResp.PartnerReferenceNo,
		AccountNo:          briResp.AccountNo,
		Name:               briResp.Name,
		Raw:                briResp,
	}

	if len(briResp.AccountInfos) > 0 {
		info := briResp.AccountInfos[0]

		if info.Amount != nil {
			modelResp.Amount = &model.Amount{
				Value:    info.Amount.Value,
				Currency: info.Amount.Currency,
			}
		}

		if info.FloatAmount != nil {
			modelResp.FloatAmount = &model.Amount{
				Value:    info.FloatAmount.Value,
				Currency: info.FloatAmount.Currency,
			}
		}

		if info.HoldAmount != nil {
			modelResp.HoldAmount = &model.Amount{
				Value:    info.HoldAmount.Value,
				Currency: info.HoldAmount.Currency,
			}
		}

		if info.AvailableBalance != nil {
			modelResp.AvailableBalance = &model.Amount{
				Value:    info.AvailableBalance.Value,
				Currency: info.AvailableBalance.Currency,
			}
		}

		if info.LedgerBalance != nil {
			modelResp.LedgerBalance = &model.Amount{
				Value:    info.LedgerBalance.Value,
				Currency: info.LedgerBalance.Currency,
			}
		}

		modelResp.Status = info.Status
	}

	if briResp.AdditionalInfo != nil {
		modelResp.ProductCode = briResp.AdditionalInfo.ProductCode
		modelResp.AccountType = briResp.AdditionalInfo.AccountType
	}

	return modelResp
}
