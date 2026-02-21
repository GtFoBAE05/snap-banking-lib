package bca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"snap-banking-lib/internal/utils"
	"snap-banking-lib/model"
)

type bcaExternalAccountInquiryAdditionalInfo struct {
	InquiryService  string        `json:"inquiryService"`
	SourceAccountNo string        `json:"sourceAccountNo,omitempty"`
	Amount          *model.Amount `json:"amount,omitempty"`
	PurposeCode     string        `json:"purposeCode,omitempty"`
}

type bcaExternalAccountInquiryRequest struct {
	BeneficiaryBankCode  string                                   `json:"beneficiaryBankCode"`
	BeneficiaryAccountNo string                                   `json:"beneficiaryAccountNo"`
	PartnerReferenceNo   string                                   `json:"partnerReferenceNo"`
	AdditionalInfo       *bcaExternalAccountInquiryAdditionalInfo `json:"additionalInfo"`
}

type bcaExternalAccountInquiryResponse struct {
	PartnerReferenceNo     string `json:"partnerReferenceNo"`
	ResponseCode           string `json:"responseCode"`
	ResponseMessage        string `json:"responseMessage"`
	ReferenceNo            string `json:"referenceNo"`
	BeneficiaryAccountName string `json:"beneficiaryAccountName"`
	BeneficiaryAccountNo   string `json:"beneficiaryAccountNo"`
	BeneficiaryBankCode    string `json:"beneficiaryBankCode"`
}

func (a *BCAAdapter) GetExternalAccountInquiry(ctx context.Context, accessToken string, request *model.ExternalAccountInquiryRequest) (*model.ExternalAccountInquiryResponse, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBCA.Name(), "GetExternalAccountInquiry")
	defer span.End()

	url, path, err := a.GetEndpoint(model.EndpointExternalAccountInquiry)
	if err != nil {
		return nil, err
	}

	requestBody := bcaExternalAccountInquiryRequest{
		BeneficiaryBankCode:  request.BeneficiaryBankCode,
		BeneficiaryAccountNo: request.BeneficiaryAccountNo,
		PartnerReferenceNo:   request.PartnerReferenceNo,
	}

	if request.AdditionalInfo != nil {
		requestBody.AdditionalInfo = &bcaExternalAccountInquiryAdditionalInfo{
			InquiryService:  request.AdditionalInfo.InquiryService,
			SourceAccountNo: request.AdditionalInfo.SourceAccountNo,
			PurposeCode:     request.AdditionalInfo.PurposeCode,
		}
		if request.AdditionalInfo.Amount != nil {
			requestBody.AdditionalInfo.Amount = &model.Amount{
				Value:    request.AdditionalInfo.Amount.Value,
				Currency: request.AdditionalInfo.Amount.Currency,
			}
		}
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, &model.ClientError{
			Operation: model.OperationMarshalExternalAccountInquiryRequest,
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
		"ORIGIN":        request.Origin,
	}

	response, err := a.httpClient.Do(ctx, "POST", url, headers, requestBodyJSON)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationExternalAccountInquiryRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		return nil, &model.NetworkError{
			Operation: model.OperationReadExternalAccountInquiryResponse,
			Err:       err,
		}
	}

	var resp bcaExternalAccountInquiryResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		span.RecordError(err)
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalExternalAccountInquiryResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessExternalAccountInquiry {
		code, message := model.MapSNAPError(resp.ResponseCode, resp.ResponseMessage)
		span.RecordError(err)
		return nil, &model.APIError{
			Code:       code,
			Message:    message,
			HTTPStatus: response.StatusCode,
			RawCode:    resp.ResponseCode,
		}
	}

	return bcaExternalAccountInquiryResponseToModel(resp), nil
}

func bcaExternalAccountInquiryResponseToModel(bcaResp bcaExternalAccountInquiryResponse) *model.ExternalAccountInquiryResponse {
	return &model.ExternalAccountInquiryResponse{
		PartnerReferenceNo:     bcaResp.PartnerReferenceNo,
		ResponseCode:           bcaResp.ResponseCode,
		ResponseMessage:        bcaResp.ResponseMessage,
		ReferenceNo:            bcaResp.ReferenceNo,
		BeneficiaryAccountName: bcaResp.BeneficiaryAccountName,
		BeneficiaryAccountNo:   bcaResp.BeneficiaryAccountNo,
		BeneficiaryBankCode:    bcaResp.BeneficiaryBankCode,
		Raw:                    bcaResp,
	}
}
