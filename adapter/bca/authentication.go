package bca

import (
	"context"
	"encoding/json"
	"io"
	"snap-banking-lib/internal/utils"
	"snap-banking-lib/model"
	"strconv"
	"time"
)

type bcaAccessTokenRequest struct {
	GrantType string `json:"grantType"`
}

type bcaAccessTokenResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`

	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   string `json:"expiresIn"`
}

func (a *BCAAdapter) GetAccessToken(ctx context.Context) (*model.AccessToken, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "GetAccessToken")
	defer span.End()

	url, _, err := a.GetEndpoint(model.EndpointAccessToken)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	requestBody := bcaAccessTokenRequest{
		GrantType: "client_credentials",
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationMarshalAccessTokenRequest,
			Err:       err,
		}
	}

	timestamp := utils.ISO8601Timestamp()

	signature, err := a.GenerateTokenSignature(ctx, timestamp)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"X-TIMESTAMP":  timestamp,
		"X-CLIENT-KEY": a.bankConfig.ClientID,
		"X-SIGNATURE":  signature,
	}

	response, err := a.httpClient.Do(ctx, "POST", url, headers, requestBodyJSON)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationAccessTokenRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadAccessTokenResponse,
			Err:       err,
		}
	}

	var resp bcaAccessTokenResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalAccessTokenResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessAccessToken {
		code, message := model.MapSNAPError(resp.ResponseCode, resp.ResponseMessage)
		span.RecordError(err)
		return nil, &model.APIError{
			Code:       code,
			Message:    message,
			HTTPStatus: response.StatusCode,
			RawCode:    resp.ResponseCode,
		}
	}

	expireIn, err := strconv.Atoi(resp.ExpiresIn)
	if err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationParseExpiresIn,
			Err:       err,
		}
	}

	token := &model.AccessToken{
		AccessToken: resp.AccessToken,
		TokenType:   resp.TokenType,
		ExpiresIn:   expireIn,
		ExpiresAt:   time.Now().Add(time.Duration(expireIn) * time.Second),
		Raw:         resp,
	}
	return token, nil
}
