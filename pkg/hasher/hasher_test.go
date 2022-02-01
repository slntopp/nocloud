package hasher

import (
	"encoding/json"
	"fmt"
	"testing"

	instances "github.com/slntopp/nocloud/pkg/instances/proto"
	services "github.com/slntopp/nocloud/pkg/services/proto"
	"google.golang.org/protobuf/proto"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

func prettyMsg(s string, msg proto.Message) {
	bt1, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		fmt.Println(s, "error:", err)
	}
	fmt.Println(s, string(bt1))
}

func initMessage() *services.Service {

	sv := structpb.NewBoolValue(true)
	svm := make(map[string]*structpb.Value)
	svm["one"] = sv

	is := instances.Instance{
		Uuid:  "Uuid",
		Title: "Title",
	}

	iss := []*instances.Instance{&is, &is}

	ig := instances.InstancesGroup{
		Uuid:      "Uuid",
		Type:      "Type",
		Config:    svm,
		Instances: iss,
		Resources: svm,
		Data:      svm,
	}

	ctx := make(map[string]*structpb.Value)
	ctx["one"] = sv
	igm := make(map[string]*instances.InstancesGroup)
	igm["one"] = &ig

	serv := &services.Service{
		Uuid:            "Uuid",
		Version:         "Version",
		Title:           "Title",
		Status:          "Status",
		Context:         ctx,
		InstancesGroups: igm,
		Hash:            "uuid",
	}

	return serv
}

func TestRedact(t *testing.T) {

	tests := []struct {
		name string
		args *services.Service
	}{
		{"test1", initMessage()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prettyMsg("Before:", tt.args)
			redact(tt.args.ProtoReflect())
			prettyMsg("After:", tt.args)
		})
	}
}

func TestGetHash(t *testing.T) {

	tests := []struct {
		name string
		args *services.Service
	}{
		{"test1", initMessage()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			v, err := GetHash(tt.args)
			if err != nil {
				t.Error(err)
				return
			}

			if v != "88e34d66e1f48b6cbac6a778e3eb9abbde4378326443503a88654bb129903f79" {

				t.Error("Non-expected ", v)
			}

		})
	}
}
