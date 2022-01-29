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
	fmt.Println(s, string(bt1))
}

func initMessage() *services.Service {

	sv0 := structpb.NewBoolValue(true)
	svm0 := make(map[string]*structpb.Value)
	svm0["zero"] = sv0

	sv1 := structpb.NewStringValue("true")
	svm1 := make(map[string]*structpb.Value)
	svm1["one"] = sv1

	is0 := instances.Instance{
		Uuid:      "Uuid",
		Title:     "Title0",
		Config:    svm0,
		Resources: svm1,
		Hash:     "Instance0",
	}

	is1 := instances.Instance{
		Uuid:      "Uuid",
		Title:     "Title1",
		Config:    svm1,
		Resources: svm0,
		Hash:     "Instance1",
	}

	iss := []*instances.Instance{&is0, &is1}

	ig := instances.InstancesGroup{
		Uuid:      "Uuid",
		Type:      "Type",
		Config:    svm0,
		Instances: iss,
		Resources: svm0,
		Data:      svm0,
		Hash:     "InstancesGroup",
	}

	ctx := make(map[string]*structpb.Value)
	ctx["one"] = sv0
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

			// msg_pm:=tt.args.ProtoReflect()
			// msg1 :=proto.Clone(msg_pm.Interface())
			// redact(msg1.ProtoReflect())
			// prettyMsg("After Orig:", tt.args)
			// prettyMsg("After Copy:", msg1)
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

			if tt.args.Hash != "116797c42e80a0cae0646baa87a83ba9f28ebb54c5a96024103f149b909b96fc" {

				t.Error("Non-expected ", tt.args.Hash)
			}

		})
	}
}
