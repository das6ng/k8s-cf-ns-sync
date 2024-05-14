package cfnssync

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListDNS(t *testing.T) {
	ctx := context.Background()
	err := InitCloudflare(ctx)
	assert.Nil(t, err)

	zid := cloudflare.ZoneIdentifier(cfZoneID)
	rec, info, err := api.ListDNSRecords(ctx, zid, cloudflare.ListDNSRecordsParams{})
	if err != nil {
		t.Fatal("err:", err.Error())
	}

	for _, r := range rec {
		t.Logf("%s  %s  %s\n", r.Type, r.Name, r.ZoneName)
	}
	// t.Log(lo.Map(rec, func(r cloudflare.DNSRecord, _ int) string {
	// 	return fmt.Sprintf("%s/%s/%s", r.Type, r.Name, r.ZoneName)
	// }))
	t.Logf("%d/%d", info.Count, info.Total)
}

func TestSyncDNS(t *testing.T) {
	ctx := context.Background()
	err := InitCloudflare(ctx)
	require.NoError(t, err)

	
}
