package cf

import (
	"context"
	"log/slog"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

type remoteStatus struct {
	zone   string
	zoneID string
	api    *cloudflare.API
	ident  *cloudflare.ResourceContainer

	records map[string]*cloudflare.DNSRecord
	expire  time.Time
}

func NewZone(ctx context.Context, zone, apiToken string) (remote *remoteStatus, err error) {
	remote = &remoteStatus{zone: zone}
	remote.api, err = cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		return
	}
	remote.zoneID, err = remote.api.ZoneIDByName(zone)
	if err != nil {
		return
	}
	remote.ident = cloudflare.ZoneIdentifier(remote.zoneID)
	return
}

func (r *remoteStatus) CheckRemote(ctx context.Context) (err error) {
	if r.expire.After(time.Now()) {
		return
	}
	records, _, err := r.api.ListDNSRecords(ctx, r.ident, cloudflare.ListDNSRecordsParams{})
	if err != nil {
		slog.ErrorContext(ctx, "ListDNSRecords fail", "err", err.Error())
		return
	}
	r.records = make(map[string]*cloudflare.DNSRecord, len(records))
	for _, rec := range records {
		r.records[rec.Name] = &rec
	}
	r.expire = time.Now().Add(5 * time.Minute)
	return
}

func (r *remoteStatus) Sync(ctx context.Context, name, content string) {
	_ = r.CheckRemote(ctx)
	rec, ok := r.records[name]
	if ok && rec.Content == content {
		slog.InfoContext(ctx, "dns record already synced with remote", "name", name, "content", content)
		return
	}

	priority := uint16(10)
	proxied := false
	const ttl = 300 // seconds
	if ok {
		// should update record content
		if _, err := r.api.UpdateDNSRecord(ctx, r.ident, cloudflare.UpdateDNSRecordParams{
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
		rec.Content = content
	} else {
		// should create new record
		rec, err := r.api.CreateDNSRecord(ctx, r.ident, cloudflare.CreateDNSRecordParams{
			Type:     "A",
			Name:     name,
			Content:  content,
			TTL:      ttl,
			Priority: &priority,
			Proxied:  &proxied,
		})
		if err != nil {
			slog.ErrorContext(ctx, "CreateDNSRecord fail", "name", name, "content", content, "err", err.Error())
			return
		}
		slog.InfoContext(ctx, "CreateDNSRecord success", "name", name, "content", content)
		r.records[rec.Name] = &rec
	}
}
