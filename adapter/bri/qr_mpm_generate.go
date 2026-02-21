package bri

import (
	"context"
	"fmt"
	"snap-banking-lib/model"
)

func (a *BRIAdapter) GenerateQRMPM(ctx context.Context, accessToken string, request *model.QRMPMGenerateRequest) (*model.QRMPMGenerateResponse, error) {
	return nil, &model.ClientError{
		Operation: model.ErrNotSupported,
		Err:       fmt.Errorf("GenerateQRMPM supported"),
	}
}
