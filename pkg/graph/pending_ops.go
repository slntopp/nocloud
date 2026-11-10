package graph

import (
	"fmt"
	"time"

	pb "github.com/slntopp/nocloud-proto/instances"
	"google.golang.org/protobuf/types/known/structpb"
)

func enqueueResourceOps(inst, oldInst *pb.Instance) {
	if inst == nil || oldInst == nil {
		return
	}
	if inst.Data == nil {
		inst.Data = map[string]*structpb.Value{}
		for k, v := range oldInst.GetData() {
			inst.Data[k] = v
		}
	}
	from := oldInst.GetResources()
	to := inst.GetResources()
	if from == nil || to == nil {
		return
	}
	var extra []*structpb.Value
	for _, field := range []string{"cpu", "ram", "drive_size", "ips_public", "ips_private"} {
		o := from[field].GetNumberValue()
		n := to[field].GetNumberValue()
		if o == n {
			continue
		}
		name := field
		switch field {
		case "drive_size":
			name = "disk_resize"
		case "ips_public", "ips_private":
			name = "ips"
		}
		st, err := structpb.NewStruct(map[string]interface{}{
			"id":     fmt.Sprintf("%s-%d", name, time.Now().UnixNano()),
			"op":     name,
			"field":  field,
			"from":   o,
			"to":     n,
			"status": "queued",
		})
		if err != nil {
			continue
		}
		extra = append(extra, structpb.NewStructValue(st))
	}
	if len(extra) == 0 {
		return
	}
	cur := inst.Data["pending_ops"].GetListValue()
	if cur == nil {
		cur = &structpb.ListValue{}
	}
	cur.Values = append(cur.Values, extra...)
	inst.Data["pending_ops"] = structpb.NewListValue(cur)
}
