package eventbus

import (
	"errors"
	"testing"
	"time"

	bpb "github.com/slntopp/nocloud-proto/billing"
	pb "github.com/slntopp/nocloud-proto/events"
	"google.golang.org/protobuf/types/known/structpb"
)

func trialPlan() []*bpb.CustomEvent {
	return []*bpb.CustomEvent{{Override: "instance_suspended", Key: aiTrialSuspendedKey}}
}

func expiry(days int, due time.Time) *pb.Event {
	data := map[string]*structpb.Value{
		"period": structpb.NewNumberValue(float64(days)),
	}
	if !due.IsZero() {
		data["next_payment_date"] = structpb.NewNumberValue(float64(due.Unix()))
	}
	return &pb.Event{Key: "expiry_notification", Data: data}
}

func TestRewriteAITrialExpiryKeepsTheDayBefore(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	event := expiry(1, now.Add(20*time.Hour))

	got, err := rewriteAITrialExpiry(event, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.GetKey() != aiTrialExpiringKey {
		t.Fatalf("key %q", got.GetKey())
	}
}

func TestRewriteAITrialExpiryDropsTheRest(t *testing.T) {
	now := time.Date(2026, 11, 21, 17, 40, 0, 0, time.UTC)
	cases := []*pb.Event{
		expiry(0, now.Add(-14*24*time.Hour)),
		expiry(1, now.Add(-14*24*time.Hour)), // "tomorrow" letter for a date already past
		expiry(3, now.Add(3*24*time.Hour)),
		expiry(7, now.Add(7*24*time.Hour)),
	}
	for _, event := range cases {
		_, err := rewriteAITrialExpiry(event, now)
		if !errors.Is(err, errEventCancelled) {
			t.Fatalf("period %v: %v", event.GetData()["period"].GetNumberValue(), err)
		}
	}
}

func TestAITrialPlan(t *testing.T) {
	if !aiTrialPlan(trialPlan()) {
		t.Fatal("plan with the suspend override is a trial plan")
	}
	if aiTrialPlan([]*bpb.CustomEvent{{Override: "instance_suspended", Key: "instance_suspended"}}) {
		t.Fatal("a plain suspend override is not a trial plan")
	}
}
