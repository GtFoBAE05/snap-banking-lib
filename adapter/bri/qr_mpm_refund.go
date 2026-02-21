package bri

import (
	"context"
	"fmt"
	"snap-banking-lib/model"
)

func (a *BRIAdapter) RefundQRMPM(ctx context.Context, accessToken string, request *model.QRMPMRefundRequest) (*model.QRMPMRefundResponse, error) {
	return nil, &model.ClientError{
		Operation: model.ErrNotSupported,
		Err:       fmt.Errorf("RefundQRMPM not supported"),
	}
}
