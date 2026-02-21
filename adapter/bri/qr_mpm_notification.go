package bri

import (
	"context"
	"fmt"
	"snap-banking-lib/model"
)

func (a *BRIAdapter) HandleQRISNotification(ctx context.Context, accessToken string, request *model.QRISNotificationRequest) (*model.QRISNotificationResponse, error) {
	return nil, &model.ClientError{
		Operation: model.ErrNotSupported,
		Err:       fmt.Errorf("HandleQRISNotification not supported"),
	}
}
