package hasher

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"

	"google.golang.org/protobuf/proto"

	pb "github.com/slntopp/nocloud/pkg/hasher/hasherpb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
)
//Delete unmarked fields from messages.
//Structs is implementation protobuf Messages in Go, and protoreflect kind of Go reflect, but it's own protobuf
func redact(msg protoreflect.Message) {
	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if proto.GetExtension(fd.Options().(*descriptorpb.FieldOptions), pb.E_Hashed).(bool) {
			//There is more nested Kinds of fields, but here considered only maps
			if fd.IsMap() && fd.MapValue().Kind() == protoreflect.MessageKind {
				//not delete inner protobuf messages like google.protobuf.Value
				if fd.MapValue().Message().FullName()[:7] == "nocloud" {
					v.Map().Range(func(km protoreflect.MapKey, vm protoreflect.Value) bool {
						redact(vm.Message())
						return true
					})
				}

			} else if fd.Kind() == protoreflect.MessageKind {
				redact(v.Message())
			}

			return true
		}
		msg.Clear(fd) //delete non-marked as E_Hashed fields
		return true
	})

}

func GetHash(msg proto.Message) (string, error) {

	redact(msg.ProtoReflect())

	bt, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return "error:", err
	}
	byteSl := md5.Sum(bt)
	return hex.EncodeToString(byteSl[:]), nil
}
