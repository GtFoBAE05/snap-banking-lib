package bri

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"snap-banking-lib/adapter"
	"snap-banking-lib/internal/crypto"
	"snap-banking-lib/internal/httpclient"
	"snap-banking-lib/internal/utils"
	"snap-banking-lib/model"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

type BRIAdapter struct {
	bankConfig *model.BankConfig
	httpClient httpclient.HttpClient
	tracer     trace.Tracer
}

func NewAdapter(bankConfig *model.BankConfig, httpClient httpclient.HttpClient, tracer trace.Tracer) (adapter.Adapter, error) {
	return &BRIAdapter{
		bankConfig: bankConfig,
		httpClient: httpClient,
		tracer:     tracer,
	}, nil
}

func (a *BRIAdapter) GetBankCode() model.BankCode {
	return model.BankBRI
}

func (a *BRIAdapter) GetBankName() string {
	return model.BankBRI.Name()
}

func (a *BRIAdapter) GetBaseURL() string {
	return a.bankConfig.APIBaseURL
}

func (a *BRIAdapter) GetPartnerBaseURL() string {
	return a.bankConfig.PartnerBaseURL
}

func (a *BRIAdapter) GetEndpoint(key model.EndpointKey) (string, string, error) {
	endpoint, ok := a.bankConfig.Endpoints[key]
	if !ok {
		return "", "", &model.ClientError{
			Operation: model.OperationGetEndpoint,
			Err:       fmt.Errorf("endpoint %s not configured", key),
		}
	}

	return a.bankConfig.APIBaseURL + endpoint.Path, endpoint.Path, nil
}

func (a *BRIAdapter) GetPartnerEndpoint(key model.EndpointKey) (string, string, error) {
	endpoint, ok := a.bankConfig.Endpoints[key]
	if !ok {
		return "", "", &model.ClientError{
			Operation: model.OperationGetEndpoint,
			Err:       fmt.Errorf("endpoint %s not configured", key),
		}
	}

	return a.bankConfig.PartnerBaseURL + endpoint.Path, endpoint.Path, nil
}

func (a *BRIAdapter) GenerateTokenSignature(ctx context.Context, timestamp string) (string, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBRI.Name(), "GenerateTokenSignature")
	defer span.End()

	stringToSign := fmt.Sprintf("%s|%s", a.bankConfig.ClientID, timestamp)

	privateKey := a.bankConfig.PrivateKey()
	if privateKey == nil {
		span.RecordError(fmt.Errorf("private key not available"))
		return "", &model.ClientError{
			Operation: model.OperationLoadPrivateKey,
			Err:       fmt.Errorf("private key not available"),
		}
	}

	signature, err := crypto.SignRSA(privateKey, []byte(stringToSign))
	if err != nil {
		span.RecordError(err)
		return "", &model.ClientError{
			Operation: model.OperationSignRequest,
			Err:       err,
		}
	}

	return signature, nil
}

func (a *BRIAdapter) GenerateServiceSignature(ctx context.Context, accessToken, method, path, timestamp, body string) (string, error) {
	ctx, span := utils.StartSpan(ctx, a.tracer, model.BankBRI.Name(), "GenerateServiceSignature")
	defer span.End()

	hash := sha256.Sum256([]byte(body))
	hashedBody := strings.ToLower(hex.EncodeToString(hash[:]))

	stringToSign := fmt.Sprintf("%s:%s:%s:%s:%s",
		method,
		path,
		accessToken,
		hashedBody,
		timestamp,
	)

	signature := crypto.SignHMAC512(a.bankConfig.ClientSecret, []byte(stringToSign))
	return signature, nil
}
