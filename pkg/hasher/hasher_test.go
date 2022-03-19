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
	bt1, err := json.MarshalIndent(msg, "", "	")
	if err != nil {
		fmt.Println(s, "error:", err)
	}
	fmt.Println(s)
	fmt.Println(string(bt1))
}

func initMessage() *services.Service {

	sv0 := structpb.NewBoolValue(true)
	svm0 := make(map[string]*structpb.Value)
	svm0["zero"] = sv0

	sv1 := structpb.NewStringValue("text")
	svm1 := make(map[string]*structpb.Value)
	svm1["one"] = sv1

	is0 := instances.Instance{
		Uuid:      "Uuid0",
		Title:     "Title0",
		Config:    svm0,
		Resources: svm1,
		// Hash:      "Instance0",
	}

	is1 := instances.Instance{
		Uuid:      "Uuid1",
		Title:     "Title1",
		Config:    svm1,
		Resources: svm0,
		// Hash:      "Instance1",
	}

	iss := []*instances.Instance{&is0, &is1}

	ig := instances.InstancesGroup{
		Uuid:      "Uuid",
		Type:      "Type",
		Config:    svm0,
		Instances: iss,
		Resources: svm0,
		Data:      svm0,
		Hash:      "InstancesGroup",
	}

	ctx := make(map[string]*structpb.Value)
	ctx["zero"] = sv0
	ctx["one"] = sv1
	igm := make(map[string]*instances.InstancesGroup)
	igm["one"] = &ig

	serv := &services.Service{
		Uuid:            "Uuid",
		Version:         "Version",
		Title:           "Title",
		Status:          "Status",
		Context:         ctx,
		InstancesGroups: igm,
		// Hash:            "uuid",
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

			SetHash(tt.args.ProtoReflect())
			prettyMsg("Result:", tt.args)

			if tt.args.Hash != "1bc6a0a8905e70c1c355da05fcbe132be8e96b54dabc447f5d847212f78b191f" {
				t.Error("Non-expected ", tt.args.Hash)
			}

		})
	}
}