package graph

import (
	"testing"

	pb "github.com/slntopp/nocloud-proto/instances"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestEnqueueResourceOpsDiskOnly(t *testing.T) {
	old := &pb.Instance{Resources: map[string]*structpb.Value{
		"cpu": structpb.NewNumberValue(1), "drive_size": structpb.NewNumberValue(20),
	}}
	inst := &pb.Instance{Resources: map[string]*structpb.Value{
		"cpu": structpb.NewNumberValue(1), "drive_size": structpb.NewNumberValue(40),
	}}
	enqueueResourceOps(inst, old)
	ops := inst.Data["pending_ops"].GetListValue().GetValues()
	if len(ops) != 1 {
		t.Fatalf("len=%d", len(ops))
	}
	if ops[0].GetStructValue().GetFields()["op"].GetStringValue() != "disk_resize" {
		t.Fatal(ops[0])
	}
}
