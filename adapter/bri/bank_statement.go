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

type briBankStatementRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo,omitempty"`
	AccountNo          string `json:"accountNo,omitempty"`
	BankCardToken      string `json:"bankCardToken,omitempty"`
	FromDateTime       string `json:"fromDateTime"`
	ToDateTime         string `json:"toDateTime"`
}

type briTotalEntries struct {
	NumberOfEntries string        `json:"numberOfEntries"`
	Amount          *model.Amount `json:"amount"`
}

type briAmountWrapper struct {
	Amount *model.Amount `json:"amount"`
}

type briDetailBalance struct {
	StartAmount []briAmountWrapper `json:"startAmount"`
	EndAmount   []briAmountWrapper `json:"endAmount"`
}

type briDetailInfo struct {
	RemarkCustom string `json:"remarkCustom"`
}

type briTransactionDetail struct {
	DetailBalance   *briDetailBalance `json:"detailBalance"`
	Amount          *model.Amount     `json:"amount"`
	TransactionDate string            `json:"transactionDate"`
	Remark          string            `json:"remark"`
	TransactionId   string            `json:"transactionId"`
	Type            string            `json:"type"`
	DetailInfo      *briDetailInfo    `json:"detailInfo"`
}

type briBankStatementResponse struct {
	ResponseCode       string                 `json:"responseCode"`
	ResponseMessage    string                 `json:"responseMessage"`
	ReferenceNo        string                 `json:"referenceNo"`
	TotalCreditEntries *briTotalEntries       `json:"totalCreditEntries"`
	TotalDebitEntries  *briTotalEntries       `json:"totalDebitEntries"`
	DetailData         []briTransactionDetail `json:"detailData"`
}

func (a *BRIAdapter) GetBankStatement(ctx context.Context, accessToken string, request *model.BankStatementRequest) (*model.BankStatementResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBRI.Name(), "GetBankStatement")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointBankStatement)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := briBankStatementRequest{
		PartnerReferenceNo: request.PartnerReferenceNo,
		AccountNo:          request.AccountNo,
		BankCardToken:      request.BankCardToken,
		FromDateTime:       request.FromDateTime.Format(time.RFC3339),
		ToDateTime:         request.ToDateTime.Format(time.RFC3339),
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalBankStatementRequest,
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
			Operation: model.OperationBankStatementRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadBankStatementResponse,
			Err:       err,
		}
	}

	var resp briBankStatementResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalBankStatementResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessBankStatement {
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

	return briBankStatementResponseToModel(resp), nil
}

func briBankStatementResponseToModel(briResp briBankStatementResponse) *model.BankStatementResponse {
	modelResp := &model.BankStatementResponse{
		ResponseCode:    briResp.ResponseCode,
		ResponseMessage: briResp.ResponseMessage,
		ReferenceNo:     briResp.ReferenceNo,
		Raw:             briResp,
	}

	if briResp.TotalCreditEntries != nil {
		modelResp.TotalCreditEntries = parseBRITotalEntries(briResp.TotalCreditEntries)
	}

	if briResp.TotalDebitEntries != nil {
		modelResp.TotalDebitEntries = parseBRITotalEntries(briResp.TotalDebitEntries)
	}

	for _, d := range briResp.DetailData {
		detail := model.TransactionDetail{
			Remark:        d.Remark,
			Type:          d.Type,
			TransactionId: d.TransactionId,
		}

		if d.Amount != nil {
			detail.Amount = &model.Amount{
				Value:    d.Amount.Value,
				Currency: d.Amount.Currency,
			}
		}

		if d.TransactionDate != "" {
			if t, err := time.Parse(time.RFC3339, d.TransactionDate); err == nil {
				detail.TransactionDate = t
			}
		}

		if d.DetailBalance != nil {
			b := model.Balance{}
			if len(d.DetailBalance.StartAmount) > 0 && d.DetailBalance.StartAmount[0].Amount != nil {
				b.StartingBalance = &model.AmountWithDateTime{
					Value:    d.DetailBalance.StartAmount[0].Amount.Value,
					Currency: d.DetailBalance.StartAmount[0].Amount.Currency,
				}
			}
			if len(d.DetailBalance.EndAmount) > 0 && d.DetailBalance.EndAmount[0].Amount != nil {
				b.EndingBalance = &model.AmountWithDateTime{
					Value:    d.DetailBalance.EndAmount[0].Amount.Value,
					Currency: d.DetailBalance.EndAmount[0].Amount.Currency,
				}
			}
			detail.Balance = &b
		}

		if d.DetailInfo != nil {
			detail.RemarkCustom = d.DetailInfo.RemarkCustom
		}

		modelResp.DetailData = append(modelResp.DetailData, detail)
	}

	return modelResp
}

func parseBRITotalEntries(e *briTotalEntries) *model.TotalEntries {
	result := &model.TotalEntries{
		NumberOfEntries: e.NumberOfEntries,
	}

	if e.Amount != nil {
		result.Amount = &model.Amount{
			Value:    e.Amount.Value,
			Currency: e.Amount.Currency,
		}
	}

	return result
}
