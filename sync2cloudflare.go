package cfnssync

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/samber/lo"
)

const (
	cfToken = "CLOUDFLARE_API_TOKEN"
	cfZone  = "CLOUDFLARE_ZONE_NAME"
)

var api *cloudflare.API

var (
	cfZoneName  string
	cfZoneID    string
	cfZoneIdent *cloudflare.ResourceContainer
)

func InitCloudflare(ctx context.Context) error {
	var err error
	token := os.Getenv(cfToken)
	api, err = cloudflare.NewWithAPIToken(token)
	if err != nil {
		return err
	}
	cfZoneName = os.Getenv(cfZone)
	cfZoneID, err = api.ZoneIDByName(cfZoneName)
	if err != nil {
		return err
	}
	cfZoneIdent = cloudflare.ZoneIdentifier(cfZoneID)
	return nil
}

func Sync2Cloudflare(ctx context.Context, name, content string) {
	slog.InfoContext(ctx, "start sync to cloudflare", "name", name, "content", content)
	if api == nil || cfZoneIdent == nil {
		slog.ErrorContext(ctx, "cloudflare api not initilized")
		return
	}
	if !strings.HasSuffix(name, cfZoneName) {
		slog.InfoContext(ctx, "zone name not matched", "expect", cfZoneName, "got", name)
		return
	}
	records, _, err := api.ListDNSRecords(ctx, cfZoneIdent, cloudflare.ListDNSRecordsParams{})
	if err != nil {
		slog.ErrorContext(ctx, "ListDNSRecords fail", "err", err.Error())
		return
	}
	rec, ok := lo.Find(records, func(rr cloudflare.DNSRecord) bool {
		return rr.Name == name
	})

	priority := uint16(10)
	proxied := false
	const ttl = 300 // seconds
	if !ok {
		// should create new record
		if _, err := api.CreateDNSRecord(ctx, cfZoneIdent, cloudflare.CreateDNSRecordParams{
			Type:     "A",
			Name:     name,
			Content:  content,
			TTL:      ttl,
			Priority: &priority,
			Proxied:  &proxied,
		}); err != nil {
			slog.ErrorContext(ctx, "CreateDNSRecord fail", "name", name, "content", content, "err", err.Error())
			return
		}
		slog.InfoContext(ctx, "CreateDNSRecord success", "name", name, "content", content)
		return
	}
	if rec.Content == content {
		slog.InfoContext(ctx, "record already exists and has the same content", "name", name, "content", content)
		return
	}
	// should update record content
	if _, err := api.UpdateDNSRecord(ctx, cfZoneIdent, cloudflare.UpdateDNSRecordParams{
		Type:    "A",
		Name:    name,
		Content: content,
		TTL:     ttl,
		Proxied: &proxied,
		ID:      rec.ID,
	}); err != nil {
		slog.ErrorContext(ctx, "UpdateDNSRecord fail", "name", name, "content", content, "err", err.Error())
		return
	}
	slog.InfoContext(ctx, "CreateDNSRecord success", "name", name, "content", content)
}
