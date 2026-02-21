package bca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"snap-banking-lib/internal/utils"
	"snap-banking-lib/model"
)

type bcaBalanceInquiryRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	AccountNo          string `json:"accountNo"`
}

type bcaAccountInfo struct {
	Amount           *model.Amount `json:"amount"`
	FloatAmount      *model.Amount `json:"floatAmount"`
	HoldAmount       *model.Amount `json:"holdAmount"`
	AvailableBalance *model.Amount `json:"availableBalance"`
}

type bcaBalanceInquiryResponse struct {
	ResponseCode       string          `json:"responseCode"`
	ResponseMessage    string          `json:"responseMessage"`
	ReferenceNo        string          `json:"referenceNo"`
	PartnerReferenceNo string          `json:"partnerReferenceNo"`
	AccountNo          string          `json:"accountNo"`
	Name               string          `json:"name"`
	AccountInfos       *bcaAccountInfo `json:"accountInfos"`
}

func (a *BCAAdapter) GetBalanceInquiry(ctx context.Context, accessToken string, request *model.BalanceInquiryRequest) (*model.BalanceInquiryResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "GetBalanceInquiry")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointBalanceInquiry)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := bcaBalanceInquiryRequest{
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

	var resp bcaBalanceInquiryResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalBalanceInquiryResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessBalanceInquiry {
		code, message := model.MapSNAPError(resp.ResponseCode, resp.ResponseMessage)
		span.RecordError(err)
		return nil, &model.APIError{
			Code:       code,
			Message:    message,
			HTTPStatus: response.StatusCode,
			RawCode:    resp.ResponseCode,
		}
	}

	return bcaResponseToModel(resp), nil
}

func bcaResponseToModel(bcaResp bcaBalanceInquiryResponse) *model.BalanceInquiryResponse {
	modelResp := &model.BalanceInquiryResponse{
		ResponseCode:       bcaResp.ResponseCode,
		ResponseMessage:    bcaResp.ResponseMessage,
		ReferenceNo:        bcaResp.ReferenceNo,
		PartnerReferenceNo: bcaResp.PartnerReferenceNo,
		AccountNo:          bcaResp.AccountNo,
		Name:               bcaResp.Name,
		Raw:                bcaResp,
	}

	if bcaResp.AccountInfos != nil {
		accountInfos := bcaResp.AccountInfos

		if accountInfos.Amount != nil {
			modelResp.Amount = &model.Amount{
				Value:    accountInfos.Amount.Value,
				Currency: accountInfos.Amount.Currency,
			}
		}

		if accountInfos.FloatAmount != nil {
			modelResp.FloatAmount = &model.Amount{
				Value:    accountInfos.FloatAmount.Value,
				Currency: accountInfos.FloatAmount.Currency,
			}
		}

		if accountInfos.HoldAmount != nil {
			modelResp.HoldAmount = &model.Amount{
				Value:    accountInfos.HoldAmount.Value,
				Currency: accountInfos.HoldAmount.Currency,
			}
		}

		if accountInfos.AvailableBalance != nil {
			modelResp.AvailableBalance = &model.Amount{
				Value:    accountInfos.AvailableBalance.Value,
				Currency: accountInfos.AvailableBalance.Currency,
			}
		}

	}

	return modelResp
}
