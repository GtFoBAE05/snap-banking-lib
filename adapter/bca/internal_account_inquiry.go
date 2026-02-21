package bca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"snap-banking-lib/internal/utils"
	"snap-banking-lib/model"
)

type bcaInternalAccountInquiryRequest struct {
	PartnerReferenceNo   string `json:"partnerReferenceNo"`
	BeneficiaryAccountNo string `json:"beneficiaryAccountNo"`
}

type bcaInternalAccountInquiryResponse struct {
	ResponseCode           string `json:"responseCode"`
	ResponseMessage        string `json:"responseMessage"`
	ReferenceNo            string `json:"referenceNo"`
	PartnerReferenceNo     string `json:"partnerReferenceNo"`
	BeneficiaryAccountName string `json:"beneficiaryAccountName"`
	BeneficiaryAccountNo   string `json:"beneficiaryAccountNo"`
}

func (a *BCAAdapter) GetInternalAccountInquiry(ctx context.Context, accessToken string, request *model.InternalAccountInquiryRequest) (*model.InternalAccountInquiryResponse, error) {
	url, path, err := a.GetEndpoint(model.EndpointInternalAccountInquiry)
	if err != nil {
		return nil, err
	}

	requestBody := bcaInternalAccountInquiryRequest{
		PartnerReferenceNo:   request.PartnerReferenceNo,
		BeneficiaryAccountNo: request.BeneficiaryAccountNo,
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, &model.ClientError{
			Operation: model.OperationMarshalInternalAccountInquiryRequest,
			Err:       err,
		}
	}

	timestamp := utils.ISO8601Timestamp()

	signature, err := a.GenerateServiceSignature(ctx, accessToken, "POST", path, timestamp, string(requestBodyJSON))
	if err != nil {
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
		return nil, &model.NetworkError{
			Operation: model.OperationInternalAccountInquiryRequest,
			Err:       err,
		}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, &model.NetworkError{
			Operation: model.OperationReadInternalAccountInquiryResponse,
			Err:       err,
		}
	}

	var resp bcaInternalAccountInquiryResponse
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		return nil, &model.ClientError{
			Operation: model.OperationUnmarshalInternalAccountInquiryResponse,
			Err:       err,
		}
	}

	if resp.ResponseCode != SuccessInternalAccountInquiry {
		code, message := model.MapSNAPError(resp.ResponseCode, resp.ResponseMessage)
		return nil, &model.APIError{
			Code:       code,
			Message:    message,
			HTTPStatus: response.StatusCode,
			RawCode:    resp.ResponseCode,
		}
	}

	return bcaInternalAccountInquiryResponseToModel(resp), nil
}

func bcaInternalAccountInquiryResponseToModel(bcaResp bcaInternalAccountInquiryResponse) *model.InternalAccountInquiryResponse {
	return &model.InternalAccountInquiryResponse{
		ResponseCode:           bcaResp.ResponseCode,
		ResponseMessage:        bcaResp.ResponseMessage,
		ReferenceNo:            bcaResp.ReferenceNo,
		PartnerReferenceNo:     bcaResp.PartnerReferenceNo,
		BeneficiaryAccountName: bcaResp.BeneficiaryAccountName,
		BeneficiaryAccountNo:   bcaResp.BeneficiaryAccountNo,
		Raw:                    bcaResp,
	}
}
