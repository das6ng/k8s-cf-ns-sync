package cfnssync

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestListDNS(t *testing.T) {
	ctx := context.Background()
	apiKey := os.Getenv("CLOUDFLARE_API_KEY")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	email := os.Getenv("CLOUDFLARE_EMAIL")
	err := InitCloudflare(context.Background(), apiKey, email)
	assert.Nil(t, err)
	// sync2Cloudflare()
	rec, info, err := api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		t.Log("err:", err.Error())
	}
	t.Log(lo.Map(rec, func(r cloudflare.DNSRecord, _ int) string {
		return fmt.Sprintf("%s/%s/%s", r.Type, r.Name, r.ZoneName)
	}))
	t.Logf("%d/%d", info.Count, info.Total)
}
