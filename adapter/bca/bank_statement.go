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

type bcaBankStatementRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	AccountNo          string `json:"accountNo"`
	FromDateTime       string `json:"fromDateTime"`
	ToDateTime         string `json:"toDateTime"`
}

type bcaAmountWithDateTime struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
	DateTime string `json:"dateTime"`
}

type bcaBalance struct {
	Amount          *bcaAmountWithDateTime `json:"amount"`
	StartingBalance *bcaAmountWithDateTime `json:"startingBalance"`
	EndingBalance   *bcaAmountWithDateTime `json:"endingBalance"`
}

type bcaTotalEntries struct {
	NumberOfEntries string        `json:"numberOfEntries"`
	Amount          *model.Amount `json:"amount"`
}

type bcaTransactionDetail struct {
	Amount          *model.Amount `json:"amount"`
	TransactionDate string        `json:"transactionDate"`
	Remark          string        `json:"remark"`
	Type            string        `json:"type"`
}

type bcaBankStatementResponse struct {
	ResponseCode       string                 `json:"responseCode"`
	ResponseMessage    string                 `json:"responseMessage"`
	ReferenceNo        string                 `json:"referenceNo"`
	PartnerReferenceNo string                 `json:"partnerReferenceNo"`
	Balance            []bcaBalance           `json:"balance"`
	TotalCreditEntries *bcaTotalEntries       `json:"totalCreditEntries"`
	TotalDebitEntries  *bcaTotalEntries       `json:"totalDebitEntries"`
	DetailData         []bcaTransactionDetail `json:"detailData"`
}

func (a *BCAAdapter) GetBankStatement(ctx context.Context, accessToken string, request *model.BankStatementRequest) (*model.BankStatementResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "GetBankStatement")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointBankStatement)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := bcaBankStatementRequest{
		PartnerReferenceNo: request.PartnerReferenceNo,
		AccountNo:          request.AccountNo,
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

	var resp bcaBankStatementResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalBankStatementResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessBankStatement {
		code, message := model.MapSNAPError(resp.ResponseCode, resp.ResponseMessage)
		span.RecordError(err)
		return nil, &model.APIError{
			Code:       code,
			Message:    message,
			HTTPStatus: response.StatusCode,
			RawCode:    resp.ResponseCode,
		}
	}

	return bcaBankStatementResponseToModel(resp), nil
}

func bcaBankStatementResponseToModel(bcaResp bcaBankStatementResponse) *model.BankStatementResponse {
	modelResp := &model.BankStatementResponse{
		ResponseCode:       bcaResp.ResponseCode,
		ResponseMessage:    bcaResp.ResponseMessage,
		ReferenceNo:        bcaResp.ReferenceNo,
		PartnerReferenceNo: bcaResp.PartnerReferenceNo,
		Raw:                bcaResp,
	}

	for _, b := range bcaResp.Balance {
		modelResp.Balance = append(modelResp.Balance, model.Balance{
			Amount:          parseAmountWithDateTime(b.Amount),
			StartingBalance: parseAmountWithDateTime(b.StartingBalance),
			EndingBalance:   parseAmountWithDateTime(b.EndingBalance),
		})
	}

	if bcaResp.TotalCreditEntries != nil {
		modelResp.TotalCreditEntries = parseTotalEntries(bcaResp.TotalCreditEntries)
	}

	if bcaResp.TotalDebitEntries != nil {
		modelResp.TotalDebitEntries = parseTotalEntries(bcaResp.TotalDebitEntries)
	}

	for _, d := range bcaResp.DetailData {
		detail := model.TransactionDetail{
			Remark: d.Remark,
			Type:   d.Type,
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

		modelResp.DetailData = append(modelResp.DetailData, detail)
	}

	return modelResp
}

func parseAmountWithDateTime(a *bcaAmountWithDateTime) *model.AmountWithDateTime {
	if a == nil {
		return nil
	}

	result := &model.AmountWithDateTime{
		Value:    a.Value,
		Currency: a.Currency,
	}

	if a.DateTime != "" {
		if t, err := time.Parse(time.RFC3339, a.DateTime); err == nil {
			result.DateTime = t
		}
	}

	return result
}

func parseTotalEntries(e *bcaTotalEntries) *model.TotalEntries {
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
