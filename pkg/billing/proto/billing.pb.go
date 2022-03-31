//
//Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.18.1
// source: pkg/billing/proto/billing.proto

package proto

import (
	proto "github.com/slntopp/nocloud/pkg/states/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Kind int32

const (
	Kind_UNSPECIFIED Kind = 0 // Shall never be used, requests will be rejected
	Kind_POSTPAID    Kind = 1 // Transaction must be processed based on End time
	Kind_PREPAID     Kind = 2 // Transaction must be processed based on Start time
)

// Enum value maps for Kind.
var (
	Kind_name = map[int32]string{
		0: "UNSPECIFIED",
		1: "POSTPAID",
		2: "PREPAID",
	}
	Kind_value = map[string]int32{
		"UNSPECIFIED": 0,
		"POSTPAID":    1,
		"PREPAID":     2,
	}
)

func (x Kind) Enum() *Kind {
	p := new(Kind)
	*p = x
	return p
}

func (x Kind) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Kind) Descriptor() protoreflect.EnumDescriptor {
	return file_pkg_billing_proto_billing_proto_enumTypes[0].Descriptor()
}

func (Kind) Type() protoreflect.EnumType {
	return &file_pkg_billing_proto_billing_proto_enumTypes[0]
}

func (x Kind) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Kind.Descriptor instead.
func (Kind) EnumDescriptor() ([]byte, []int) {
	return file_pkg_billing_proto_billing_proto_rawDescGZIP(), []int{0}
}

type Plan struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Title string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	// Map of resource keys to their configurations
	// Resouce key is one of driver supported keys, like cpu, ram, etc.
	Resources map[string]*ResourceConf `protobuf:"bytes,2,rep,name=resources,proto3" json:"resources,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Plan) Reset() {
	*x = Plan{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_billing_proto_billing_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Plan) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Plan) ProtoMessage() {}

func (x *Plan) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_billing_proto_billing_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Plan.ProtoReflect.Descriptor instead.
func (*Plan) Descriptor() ([]byte, []int) {
	return file_pkg_billing_proto_billing_proto_rawDescGZIP(), []int{0}
}

func (x *Plan) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Plan) GetResources() map[string]*ResourceConf {
	if x != nil {
		return x.Resources
	}
	return nil
}

type ResourceConf struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Kind   Kind    `protobuf:"varint,1,opt,name=kind,proto3,enum=nocloud.billing.Kind" json:"kind,omitempty"`
	Price  float32 `protobuf:"fixed32,2,opt,name=price,proto3" json:"price,omitempty"`
	Period uint64  `protobuf:"varint,3,opt,name=period,proto3" json:"period,omitempty"` // if set to 0, then it's a one-time payment
	// If except set to true then transaction will be created if Instance is in one of the states listed in on
	// If except set to false then transaction will be created if Instance is NOT in one of the states listed in on
	Except bool                 `protobuf:"varint,4,opt,name=except,proto3" json:"except,omitempty"`
	On     []proto.NoCloudState `protobuf:"varint,5,rep,packed,name=on,proto3,enum=nocloud.states.NoCloudState" json:"on,omitempty"`
}

func (x *ResourceConf) Reset() {
	*x = ResourceConf{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_billing_proto_billing_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceConf) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceConf) ProtoMessage() {}

func (x *ResourceConf) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_billing_proto_billing_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceConf.ProtoReflect.Descriptor instead.
func (*ResourceConf) Descriptor() ([]byte, []int) {
	return file_pkg_billing_proto_billing_proto_rawDescGZIP(), []int{1}
}

func (x *ResourceConf) GetKind() Kind {
	if x != nil {
		return x.Kind
	}
	return Kind_UNSPECIFIED
}

func (x *ResourceConf) GetPrice() float32 {
	if x != nil {
		return x.Price
	}
	return 0
}

func (x *ResourceConf) GetPeriod() uint64 {
	if x != nil {
		return x.Period
	}
	return 0
}

func (x *ResourceConf) GetExcept() bool {
	if x != nil {
		return x.Except
	}
	return false
}

func (x *ResourceConf) GetOn() []proto.NoCloudState {
	if x != nil {
		return x.On
	}
	return nil
}

type Transaction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uuid      string   `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`            // transaction UUID
	Exec      uint64   `protobuf:"varint,2,opt,name=exec,proto3" json:"exec,omitempty"`           // Timestamp of Transaction planned execution time
	Proc      uint64   `protobuf:"varint,3,opt,name=proc,proto3" json:"proc,omitempty"`           // Timestamp of Transaction processing time
	Processed bool     `protobuf:"varint,4,opt,name=processed,proto3" json:"processed,omitempty"` // Wether Transaction has been processed(applied to Account balance, etc)
	Account   string   `protobuf:"bytes,5,opt,name=account,proto3" json:"account,omitempty"`
	Service   string   `protobuf:"bytes,6,opt,name=service,proto3" json:"service,omitempty"`
	Records   []string `protobuf:"bytes,7,rep,name=records,proto3" json:"records,omitempty"` // list of records UUIDs
	Total     float32  `protobuf:"fixed32,8,opt,name=total,proto3" json:"total,omitempty"`   // Transaction total value in NCU
	// Transaction meta data, like
	// meta: {
	// total: <number> // resource "quantity", e.g. CPU cores, RAM Mb, Drive Mb, IP quantity
	// price_atm: <number> // hourly price per quant of resouce at the moment, e.g. 1 NCU
	// [other keys]: <any> // for example Drive Type(SSD/HDD/NVMe/etc)
	// }
	Meta map[string]*structpb.Value `protobuf:"bytes,9,rep,name=meta,proto3" json:"meta,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Transaction) Reset() {
	*x = Transaction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_billing_proto_billing_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Transaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transaction) ProtoMessage() {}

func (x *Transaction) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_billing_proto_billing_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transaction.ProtoReflect.Descriptor instead.
func (*Transaction) Descriptor() ([]byte, []int) {
	return file_pkg_billing_proto_billing_proto_rawDescGZIP(), []int{2}
}

func (x *Transaction) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *Transaction) GetExec() uint64 {
	if x != nil {
		return x.Exec
	}
	return 0
}

func (x *Transaction) GetProc() uint64 {
	if x != nil {
		return x.Proc
	}
	return 0
}

func (x *Transaction) GetProcessed() bool {
	if x != nil {
		return x.Processed
	}
	return false
}

func (x *Transaction) GetAccount() string {
	if x != nil {
		return x.Account
	}
	return ""
}

func (x *Transaction) GetService() string {
	if x != nil {
		return x.Service
	}
	return ""
}

func (x *Transaction) GetRecords() []string {
	if x != nil {
		return x.Records
	}
	return nil
}

func (x *Transaction) GetTotal() float32 {
	if x != nil {
		return x.Total
	}
	return 0
}

func (x *Transaction) GetMeta() map[string]*structpb.Value {
	if x != nil {
		return x.Meta
	}
	return nil
}

type Record struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uuid      string  `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`            // Record UUID
	Start     uint64  `protobuf:"varint,2,opt,name=start,proto3" json:"start,omitempty"`         // Timestamp of Record coverage frame start
	End       uint64  `protobuf:"varint,3,opt,name=end,proto3" json:"end,omitempty"`             // Timestamp of Record coverage frame end
	Proc      uint64  `protobuf:"varint,4,opt,name=proc,proto3" json:"proc,omitempty"`           // Timestamp of Record processing time
	Processed bool    `protobuf:"varint,5,opt,name=processed,proto3" json:"processed,omitempty"` // Wether Record has been processed(converted to Transaction)
	Instance  string  `protobuf:"bytes,6,opt,name=instance,proto3" json:"instance,omitempty"`    // Instance UUID
	Resource  string  `protobuf:"bytes,7,opt,name=resource,proto3" json:"resource,omitempty"`    // Resource key
	Total     float32 `protobuf:"fixed32,8,opt,name=total,proto3" json:"total,omitempty"`        // Record total value in NCU
	// Record meta data, like
	// meta: {
	// total: <number> // resource "quantity", e.g. CPU cores, RAM Mb, Drive Mb, IP quantity
	// price_atm: <number> // hourly price per quant of resouce at the moment, e.g. 1 NCU
	// [other keys]: <any> // for example Drive Type(SSD/HDD/NVMe/etc)
	// }
	Meta map[string]*structpb.Value `protobuf:"bytes,9,rep,name=meta,proto3" json:"meta,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Record) Reset() {
	*x = Record{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_billing_proto_billing_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Record) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Record) ProtoMessage() {}

func (x *Record) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_billing_proto_billing_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Record.ProtoReflect.Descriptor instead.
func (*Record) Descriptor() ([]byte, []int) {
	return file_pkg_billing_proto_billing_proto_rawDescGZIP(), []int{3}
}

func (x *Record) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *Record) GetStart() uint64 {
	if x != nil {
		return x.Start
	}
	return 0
}

func (x *Record) GetEnd() uint64 {
	if x != nil {
		return x.End
	}
	return 0
}

func (x *Record) GetProc() uint64 {
	if x != nil {
		return x.Proc
	}
	return 0
}

func (x *Record) GetProcessed() bool {
	if x != nil {
		return x.Processed
	}
	return false
}

func (x *Record) GetInstance() string {
	if x != nil {
		return x.Instance
	}
	return ""
}

func (x *Record) GetResource() string {
	if x != nil {
		return x.Resource
	}
	return ""
}

func (x *Record) GetTotal() float32 {
	if x != nil {
		return x.Total
	}
	return 0
}

func (x *Record) GetMeta() map[string]*structpb.Value {
	if x != nil {
		return x.Meta
	}
	return nil
}

var File_pkg_billing_proto_billing_proto protoreflect.FileDescriptor

var file_pkg_billing_proto_billing_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x70, 0x6b, 0x67, 0x2f, 0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x0f, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x62, 0x69, 0x6c, 0x6c, 0x69,
	0x6e, 0x67, 0x1a, 0x1d, 0x70, 0x6b, 0x67, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x73, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0xbd, 0x01, 0x0a, 0x04, 0x50, 0x6c, 0x61, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x42,
	0x0a, 0x09, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x24, 0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x62, 0x69, 0x6c, 0x6c,
	0x69, 0x6e, 0x67, 0x2e, 0x50, 0x6c, 0x61, 0x6e, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x09, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x73, 0x1a, 0x5b, 0x0a, 0x0e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x33, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e,
	0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x43, 0x6f, 0x6e, 0x66, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22,
	0xad, 0x01, 0x0a, 0x0c, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x66,
	0x12, 0x29, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x15,
	0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67,
	0x2e, 0x4b, 0x69, 0x6e, 0x64, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x70,
	0x72, 0x69, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x02, 0x52, 0x05, 0x70, 0x72, 0x69, 0x63,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x06, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x78, 0x63,
	0x65, 0x70, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x65, 0x78, 0x63, 0x65, 0x70,
	0x74, 0x12, 0x2c, 0x0a, 0x02, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x1c, 0x2e,
	0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x73, 0x2e, 0x4e,
	0x6f, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x02, 0x6f, 0x6e, 0x22,
	0xd8, 0x02, 0x0a, 0x0b, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x75,
	0x75, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x65, 0x78, 0x65, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x04, 0x65, 0x78, 0x65, 0x63, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x72, 0x6f, 0x63, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x04, 0x70, 0x72, 0x6f, 0x63, 0x12, 0x1c, 0x0a, 0x09, 0x70,
	0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09,
	0x70, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x65, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x63, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x18, 0x0a,
	0x07, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07,
	0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x02, 0x52, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x12, 0x3a, 0x0a,
	0x04, 0x6d, 0x65, 0x74, 0x61, 0x18, 0x09, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x6e, 0x6f,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x2e, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x1a, 0x4f, 0x0a, 0x09, 0x4d, 0x65, 0x74,
	0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xcc, 0x02, 0x0a, 0x06, 0x52,
	0x65, 0x63, 0x6f, 0x72, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12,
	0x10, 0x0a, 0x03, 0x65, 0x6e, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x65, 0x6e,
	0x64, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x72, 0x6f, 0x63, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x04, 0x70, 0x72, 0x6f, 0x63, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73,
	0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73,
	0x73, 0x65, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x12,
	0x1a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x74,
	0x6f, 0x74, 0x61, 0x6c, 0x18, 0x08, 0x20, 0x01, 0x28, 0x02, 0x52, 0x05, 0x74, 0x6f, 0x74, 0x61,
	0x6c, 0x12, 0x35, 0x0a, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x18, 0x09, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x21, 0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e,
	0x67, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x1a, 0x4f, 0x0a, 0x09, 0x4d, 0x65, 0x74, 0x61,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x2a, 0x32, 0x0a, 0x04, 0x4b, 0x69, 0x6e,
	0x64, 0x12, 0x0f, 0x0a, 0x0b, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44,
	0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x50, 0x4f, 0x53, 0x54, 0x50, 0x41, 0x49, 0x44, 0x10, 0x01,
	0x12, 0x0b, 0x0a, 0x07, 0x50, 0x52, 0x45, 0x50, 0x41, 0x49, 0x44, 0x10, 0x02, 0x42, 0xae, 0x01,
	0x0a, 0x13, 0x63, 0x6f, 0x6d, 0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x62, 0x69,
	0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x42, 0x0c, 0x42, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x73, 0x6c, 0x6e, 0x74, 0x6f, 0x70, 0x70, 0x2f, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0xa2, 0x02, 0x03, 0x4e, 0x42, 0x58, 0xaa, 0x02, 0x0f, 0x4e, 0x6f, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x2e, 0x42, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0xca, 0x02, 0x0f, 0x4e, 0x6f,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x5c, 0x42, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0xe2, 0x02, 0x1b,
	0x4e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x5c, 0x42, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x5c,
	0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x10, 0x4e, 0x6f,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x3a, 0x3a, 0x42, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_billing_proto_billing_proto_rawDescOnce sync.Once
	file_pkg_billing_proto_billing_proto_rawDescData = file_pkg_billing_proto_billing_proto_rawDesc
)

func file_pkg_billing_proto_billing_proto_rawDescGZIP() []byte {
	file_pkg_billing_proto_billing_proto_rawDescOnce.Do(func() {
		file_pkg_billing_proto_billing_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_billing_proto_billing_proto_rawDescData)
	})
	return file_pkg_billing_proto_billing_proto_rawDescData
}

var file_pkg_billing_proto_billing_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_pkg_billing_proto_billing_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_pkg_billing_proto_billing_proto_goTypes = []interface{}{
	(Kind)(0),               // 0: nocloud.billing.Kind
	(*Plan)(nil),            // 1: nocloud.billing.Plan
	(*ResourceConf)(nil),    // 2: nocloud.billing.ResourceConf
	(*Transaction)(nil),     // 3: nocloud.billing.Transaction
	(*Record)(nil),          // 4: nocloud.billing.Record
	nil,                     // 5: nocloud.billing.Plan.ResourcesEntry
	nil,                     // 6: nocloud.billing.Transaction.MetaEntry
	nil,                     // 7: nocloud.billing.Record.MetaEntry
	(proto.NoCloudState)(0), // 8: nocloud.states.NoCloudState
	(*structpb.Value)(nil),  // 9: google.protobuf.Value
}
var file_pkg_billing_proto_billing_proto_depIdxs = []int32{
	5, // 0: nocloud.billing.Plan.resources:type_name -> nocloud.billing.Plan.ResourcesEntry
	0, // 1: nocloud.billing.ResourceConf.kind:type_name -> nocloud.billing.Kind
	8, // 2: nocloud.billing.ResourceConf.on:type_name -> nocloud.states.NoCloudState
	6, // 3: nocloud.billing.Transaction.meta:type_name -> nocloud.billing.Transaction.MetaEntry
	7, // 4: nocloud.billing.Record.meta:type_name -> nocloud.billing.Record.MetaEntry
	2, // 5: nocloud.billing.Plan.ResourcesEntry.value:type_name -> nocloud.billing.ResourceConf
	9, // 6: nocloud.billing.Transaction.MetaEntry.value:type_name -> google.protobuf.Value
	9, // 7: nocloud.billing.Record.MetaEntry.value:type_name -> google.protobuf.Value
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_pkg_billing_proto_billing_proto_init() }
func file_pkg_billing_proto_billing_proto_init() {
	if File_pkg_billing_proto_billing_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_billing_proto_billing_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Plan); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_billing_proto_billing_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceConf); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_billing_proto_billing_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Transaction); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_billing_proto_billing_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Record); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_billing_proto_billing_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_billing_proto_billing_proto_goTypes,
		DependencyIndexes: file_pkg_billing_proto_billing_proto_depIdxs,
		EnumInfos:         file_pkg_billing_proto_billing_proto_enumTypes,
		MessageInfos:      file_pkg_billing_proto_billing_proto_msgTypes,
	}.Build()
	File_pkg_billing_proto_billing_proto = out.File
	file_pkg_billing_proto_billing_proto_rawDesc = nil
	file_pkg_billing_proto_billing_proto_goTypes = nil
	file_pkg_billing_proto_billing_proto_depIdxs = nil
}
