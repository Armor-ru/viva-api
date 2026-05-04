package viva_api

import (
	"context"

	"github.com/Armor-ru/viva-api/internal/vivaclient"
)

type PartnerSubscriptionAPI interface {
	GetSubscriberInfo(ctx context.Context, msisdn string) (*vivaclient.GetSubInfoResponse, error)
	InitSubscription(ctx context.Context, phoneNum, productName string) (*vivaclient.ResponseModel, error)
	ConfirmSubscription(ctx context.Context, phoneNum, productName string, otp *string) (*vivaclient.ResponseModel, error)
}
