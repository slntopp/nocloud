package hasher

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Pallinder/go-randomdata"
	billing "github.com/slntopp/nocloud/pkg/billing/proto"
	instances "github.com/slntopp/nocloud/pkg/instances/proto"
	services "github.com/slntopp/nocloud/pkg/services/proto"
	states "github.com/slntopp/nocloud/pkg/states/proto"
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
	igm := []*instances.InstancesGroup{&ig}

	serv := &services.Service{
		Uuid:            "Uuid",
		Version:         "Version",
		Title:           "Title",
		Status:          services.ServiceStatus_INIT,
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

// Must be empty hash, enum shouldn't cause panic
func TestRedactWithBP(t *testing.T) {
	i := instances.Instance{
		Title: randomdata.SillyName(),
		Config: map[string]*structpb.Value{},
	}

	redact(i.ProtoReflect())
	t.Log(i.Hash)

	hash := i.Hash

	i.BillingPlan = &billing.Plan{
		Title: randomdata.SillyName(),
		Type: "whatever",
		Resources: []*billing.ResourceConf{
			{
				On: []states.NoCloudState{
					3, 4, 5,
				},
			},
		},
	}
	redact(i.ProtoReflect())
	t.Log(i.Hash)

	if hash != i.Hash {
		t.Fatalf("Expected %s, got %s", hash, i.Hash)
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

			if tt.args.Hash != "f4a0f98677c86ef04f438a951ef6e5b0a1e844fe8311bf785990b19bdbf3b8fe" {
				t.Error("Non-expected ", tt.args.Hash)
			}

		})
	}
}