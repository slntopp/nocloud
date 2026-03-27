package ksef

const KSEF_EVENTS = "billing"

const (
	KsefSyncEnqueued = "ksef_sync_enqueued"
)

func Topic(key string) string {
	return KSEF_EVENTS + "." + key
}
