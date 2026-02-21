package bri

import (
	"context"
	"fmt"
	"snap-banking-lib/model"
)

func (a *BRIAdapter) GetQRMPMInquiry(ctx context.Context, accessToken string, request *model.QRMPMInquiryRequest) (*model.QRMPMInquiryResponse, error) {
	return nil, &model.ClientError{
		Operation: model.ErrNotSupported,
		Err:       fmt.Errorf("GetQRMPMInquiry not supported"),
	}
}
