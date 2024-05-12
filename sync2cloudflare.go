package cfnssync

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
)

const (
	cfAPI  = "CLOUDFLARE_API_KEY"
	cfMail = "CLOUDFLARE_API_EMAIL"
)

var api *cloudflare.API

func InitCloudflare(ctx context.Context, apiKey, email string) error {
	var err error
	api, err = cloudflare.NewWithAPIToken(apiKey)
	if err != nil {
		return err
	}
	return nil
}

func sync2Cloudflare(ctx context.Context, name, val string) {
	api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(""), cloudflare.ListDNSRecordsParams{})
}
