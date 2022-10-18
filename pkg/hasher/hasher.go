/*
Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package hasher

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"google.golang.org/protobuf/proto"

	pb "github.com/slntopp/nocloud/pkg/hasher/hasherpb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
)

// Delete unmarked fields from messages.
// Structs is a implementation protobuf Messages in Go, and protoreflect like Go-reflect
func redact(msg protoreflect.Message) (save4hash bool) {
	save4hash = false
	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {

		if proto.GetExtension(fd.Options().(*descriptorpb.FieldOptions), pb.E_Skipped).(bool) {
			return false
		}

		if proto.GetExtension(fd.Options().(*descriptorpb.FieldOptions), pb.E_Hashed).(bool) {
			save4hash = true
			return true
		}

		save4hashChild := false

		//check and not delete inner protobuf messages like google.protobuf.Value
		checkKin := func(s protoreflect.FullName) bool {
			if len(s) > 7 && s[:7] == "nocloud" {
				return true
			}
			return false
		}

		//There is more nested Kinds of fields, but here considered only maps and lists
		if fd.IsMap() {
			if fd.MapValue().Kind() == protoreflect.MessageKind && checkKin(fd.MapValue().Message().FullName()) {
				v.Map().Range(func(km protoreflect.MapKey, vm protoreflect.Value) bool {
					if redact(vm.Message()) {
						save4hashChild = true
					}
					return true
				})
			}
		} else if fd.IsList() {
			for list, i := v.List(), 0; i < list.Len(); i++ {
				if redact(list.Get(i).Message()) {
					save4hashChild = true
				}

			}
		} else if fd.Kind() == protoreflect.MessageKind && checkKin(fd.Message().FullName()) {
			if redact(v.Message()) {
				save4hashChild = true
			}
		}

		if !save4hashChild {
			msg.Clear(fd) //delete non-marked as E_Hashed fields
		} else {
			save4hash = true
		}

		return true
	})

	return save4hash
}

// Making own copy of struct msg, cleaning it and returning hash
func getHash(msg protoreflect.Message) (string, error) {

	msg_proto := proto.Clone(msg.Interface()).ProtoReflect()
	redact(msg_proto)

	bt, err := json.MarshalIndent(msg_proto.Interface(), "", "  ")
	if err != nil {
		return "error:", err
	}
	byteSl := sha256.Sum256(bt)
	return hex.EncodeToString(byteSl[:]), nil
}

// Setting hash values in hash fields
func SetHash(msg protoreflect.Message) (err_sh error) {
	err_sh = nil

	//initialize blank hash fields
	//msg.Range don't see blank fields
	md := msg.Descriptor()
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if !msg.Has(fd) && proto.GetExtension(fd.Options().(*descriptorpb.FieldOptions), pb.E_Hash).(bool) {
			msg.Set(fd, protoreflect.ValueOfString("init"))
		}
	}

	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {

		if proto.GetExtension(fd.Options().(*descriptorpb.FieldOptions), pb.E_Skipped).(bool) {
			return false
		}

		if fd.IsMap() {
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				v.Map().Range(func(km protoreflect.MapKey, vm protoreflect.Value) bool {
					err := SetHash(vm.Message())
					if err != nil {
						err_sh = err
						return false
					}
					return true
				})

				if err_sh != nil {
					return false
				}
			}
		} else if fd.IsList() {
			for list, i := v.List(), 0; i < list.Len(); i++ {
				err := SetHash(list.Get(i).Message())
				if err != nil {
					err_sh = err
					return false
				}
			}
		} else if fd.Kind() == protoreflect.MessageKind {
			err := SetHash(v.Message())
			if err != nil {
				err_sh = err
				return false
			}
		}

		if proto.GetExtension(fd.Options().(*descriptorpb.FieldOptions), pb.E_Hash).(bool) {
			hash, err := getHash(msg)
			if err != nil {
				err_sh = err
				return false
			}

			msg.Set(fd, protoreflect.ValueOfString(hash))
		}

		return true
	})

	return err_sh

}
