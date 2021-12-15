//
//Copyright © 2021 Nikita Ivanovski info@slnt-opp.xyz
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
// source: pkg/drivers/instance/vanilla/driver.proto

package vanilla

import (
	proto "github.com/slntopp/nocloud/pkg/instances/proto"
	proto1 "github.com/slntopp/nocloud/pkg/services_providers/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetTypeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetTypeRequest) Reset() {
	*x = GetTypeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTypeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTypeRequest) ProtoMessage() {}

func (x *GetTypeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTypeRequest.ProtoReflect.Descriptor instead.
func (*GetTypeRequest) Descriptor() ([]byte, []int) {
	return file_pkg_drivers_instance_vanilla_driver_proto_rawDescGZIP(), []int{0}
}

type GetTypeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
}

func (x *GetTypeResponse) Reset() {
	*x = GetTypeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTypeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTypeResponse) ProtoMessage() {}

func (x *GetTypeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTypeResponse.ProtoReflect.Descriptor instead.
func (*GetTypeResponse) Descriptor() ([]byte, []int) {
	return file_pkg_drivers_instance_vanilla_driver_proto_rawDescGZIP(), []int{1}
}

func (x *GetTypeResponse) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

type UpRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Group            *proto.InstancesGroup    `protobuf:"bytes,1,opt,name=group,proto3" json:"group,omitempty"`
	ServicesProvider *proto1.ServicesProvider `protobuf:"bytes,2,opt,name=services_provider,json=servicesProvider,proto3" json:"services_provider,omitempty"`
}

func (x *UpRequest) Reset() {
	*x = UpRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpRequest) ProtoMessage() {}

func (x *UpRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpRequest.ProtoReflect.Descriptor instead.
func (*UpRequest) Descriptor() ([]byte, []int) {
	return file_pkg_drivers_instance_vanilla_driver_proto_rawDescGZIP(), []int{2}
}

func (x *UpRequest) GetGroup() *proto.InstancesGroup {
	if x != nil {
		return x.Group
	}
	return nil
}

func (x *UpRequest) GetServicesProvider() *proto1.ServicesProvider {
	if x != nil {
		return x.ServicesProvider
	}
	return nil
}

type UpResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Group *proto.InstancesGroup `protobuf:"bytes,1,opt,name=group,proto3" json:"group,omitempty"`
}

func (x *UpResponse) Reset() {
	*x = UpResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpResponse) ProtoMessage() {}

func (x *UpResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpResponse.ProtoReflect.Descriptor instead.
func (*UpResponse) Descriptor() ([]byte, []int) {
	return file_pkg_drivers_instance_vanilla_driver_proto_rawDescGZIP(), []int{3}
}

func (x *UpResponse) GetGroup() *proto.InstancesGroup {
	if x != nil {
		return x.Group
	}
	return nil
}

type DownRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Group            *proto.InstancesGroup    `protobuf:"bytes,1,opt,name=group,proto3" json:"group,omitempty"`
	ServicesProvider *proto1.ServicesProvider `protobuf:"bytes,2,opt,name=services_provider,json=servicesProvider,proto3" json:"services_provider,omitempty"`
}

func (x *DownRequest) Reset() {
	*x = DownRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownRequest) ProtoMessage() {}

func (x *DownRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownRequest.ProtoReflect.Descriptor instead.
func (*DownRequest) Descriptor() ([]byte, []int) {
	return file_pkg_drivers_instance_vanilla_driver_proto_rawDescGZIP(), []int{4}
}

func (x *DownRequest) GetGroup() *proto.InstancesGroup {
	if x != nil {
		return x.Group
	}
	return nil
}

func (x *DownRequest) GetServicesProvider() *proto1.ServicesProvider {
	if x != nil {
		return x.ServicesProvider
	}
	return nil
}

type DownResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DownResponse) Reset() {
	*x = DownResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownResponse) ProtoMessage() {}

func (x *DownResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownResponse.ProtoReflect.Descriptor instead.
func (*DownResponse) Descriptor() ([]byte, []int) {
	return file_pkg_drivers_instance_vanilla_driver_proto_rawDescGZIP(), []int{5}
}

// Available only for users with SUPERUSER and ADMIN access to platform namespace
type ActionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServicesProvider *proto1.ServicesProvider `protobuf:"bytes,1,opt,name=services_provider,json=servicesProvider,proto3" json:"services_provider,omitempty"`
	Action           string                   `protobuf:"bytes,2,opt,name=action,proto3" json:"action,omitempty"`
	Params           []*anypb.Any             `protobuf:"bytes,3,rep,name=params,proto3" json:"params,omitempty"`
}

func (x *ActionRequest) Reset() {
	*x = ActionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ActionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ActionRequest) ProtoMessage() {}

func (x *ActionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ActionRequest.ProtoReflect.Descriptor instead.
func (*ActionRequest) Descriptor() ([]byte, []int) {
	return file_pkg_drivers_instance_vanilla_driver_proto_rawDescGZIP(), []int{6}
}

func (x *ActionRequest) GetServicesProvider() *proto1.ServicesProvider {
	if x != nil {
		return x.ServicesProvider
	}
	return nil
}

func (x *ActionRequest) GetAction() string {
	if x != nil {
		return x.Action
	}
	return ""
}

func (x *ActionRequest) GetParams() []*anypb.Any {
	if x != nil {
		return x.Params
	}
	return nil
}

var File_pkg_drivers_instance_vanilla_driver_proto protoreflect.FileDescriptor

var file_pkg_drivers_instance_vanilla_driver_proto_rawDesc = []byte{
	0x0a, 0x29, 0x70, 0x6b, 0x67, 0x2f, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x69, 0x6e,
	0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x76, 0x61, 0x6e, 0x69, 0x6c, 0x6c, 0x61, 0x2f, 0x64,
	0x72, 0x69, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1f, 0x6e, 0x6f, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x64, 0x72,
	0x69, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x61, 0x6e, 0x69, 0x6c, 0x6c, 0x61, 0x1a, 0x19, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e,
	0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x23, 0x70, 0x6b, 0x67, 0x2f, 0x69, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x69, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x35, 0x70, 0x6b,
	0x67, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x5f, 0x70, 0x72, 0x6f, 0x76, 0x69,
	0x64, 0x65, 0x72, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x73, 0x5f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x10, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x25, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0x9f, 0x01, 0x0a,
	0x09, 0x55, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x37, 0x0a, 0x05, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x6e, 0x6f, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x2e, 0x49, 0x6e,
	0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x52, 0x05, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x12, 0x59, 0x0a, 0x11, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x5f,
	0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2c,
	0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x73, 0x5f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x73, 0x2e, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x73, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x52, 0x10, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x22, 0x45,
	0x0a, 0x0a, 0x55, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x37, 0x0a, 0x05,
	0x67, 0x72, 0x6f, 0x75, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x6e, 0x6f,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x2e,
	0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x52, 0x05,
	0x67, 0x72, 0x6f, 0x75, 0x70, 0x22, 0xa1, 0x01, 0x0a, 0x0b, 0x44, 0x6f, 0x77, 0x6e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x37, 0x0a, 0x05, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x69,
	0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x2e, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x73, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x52, 0x05, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x12, 0x59,
	0x0a, 0x11, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x5f, 0x70, 0x72, 0x6f, 0x76, 0x69,
	0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x6e, 0x6f, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x5f, 0x70, 0x72, 0x6f,
	0x76, 0x69, 0x64, 0x65, 0x72, 0x73, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x50,
	0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x52, 0x10, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x73, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x22, 0x0e, 0x0a, 0x0c, 0x44, 0x6f, 0x77,
	0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0xb0, 0x01, 0x0a, 0x0d, 0x41, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x59, 0x0a, 0x11, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x5f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64,
	0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x5f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64,
	0x65, 0x72, 0x73, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x50, 0x72, 0x6f, 0x76,
	0x69, 0x64, 0x65, 0x72, 0x52, 0x10, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x50, 0x72,
	0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2c,
	0x0a, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x41, 0x6e, 0x79, 0x52, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x32, 0x8c, 0x05, 0x0a,
	0x0d, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x73,
	0x0a, 0x19, 0x54, 0x65, 0x73, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f,
	0x76, 0x69, 0x64, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x2c, 0x2e, 0x6e, 0x6f,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x5f, 0x70,
	0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x73, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x73, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x1a, 0x28, 0x2e, 0x6e, 0x6f, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x5f, 0x70, 0x72, 0x6f,
	0x76, 0x69, 0x64, 0x65, 0x72, 0x73, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x83, 0x01, 0x0a, 0x18, 0x54, 0x65, 0x73, 0x74, 0x49, 0x6e, 0x73, 0x74,
	0x61, 0x6e, 0x63, 0x65, 0x73, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x12, 0x32, 0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x73, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x73, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x33, 0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x69,
	0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x49, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x6c, 0x0a, 0x07, 0x47, 0x65, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x2f, 0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x69,
	0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e, 0x76,
	0x61, 0x6e, 0x69, 0x6c, 0x6c, 0x61, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x30, 0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e,
	0x76, 0x61, 0x6e, 0x69, 0x6c, 0x6c, 0x61, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x5d, 0x0a, 0x02, 0x55, 0x70, 0x12, 0x2a, 0x2e,
	0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65,
	0x2e, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x61, 0x6e, 0x69, 0x6c, 0x6c, 0x61, 0x2e,
	0x55, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2b, 0x2e, 0x6e, 0x6f, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x64, 0x72, 0x69,
	0x76, 0x65, 0x72, 0x2e, 0x76, 0x61, 0x6e, 0x69, 0x6c, 0x6c, 0x61, 0x2e, 0x55, 0x70, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x63, 0x0a, 0x04, 0x44, 0x6f, 0x77, 0x6e, 0x12, 0x2c,
	0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x2e, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x61, 0x6e, 0x69, 0x6c, 0x6c, 0x61,
	0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2d, 0x2e, 0x6e,
	0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x2e,
	0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x61, 0x6e, 0x69, 0x6c, 0x6c, 0x61, 0x2e, 0x44,
	0x6f, 0x77, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x4e, 0x0a, 0x06, 0x49,
	0x6e, 0x76, 0x6f, 0x6b, 0x65, 0x12, 0x2e, 0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e,
	0x76, 0x61, 0x6e, 0x69, 0x6c, 0x6c, 0x61, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x42, 0x8b, 0x02, 0x0a, 0x23,
	0x63, 0x6f, 0x6d, 0x2e, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x69, 0x6e, 0x73, 0x74,
	0x61, 0x6e, 0x63, 0x65, 0x2e, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x61, 0x6e, 0x69,
	0x6c, 0x6c, 0x61, 0x42, 0x0b, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x50, 0x01, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73,
	0x6c, 0x6e, 0x74, 0x6f, 0x70, 0x70, 0x2f, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x69, 0x6e, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x2f, 0x76, 0x61, 0x6e, 0x69, 0x6c, 0x6c, 0x61, 0xa2, 0x02, 0x04, 0x4e, 0x49,
	0x44, 0x56, 0xaa, 0x02, 0x1f, 0x4e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x49, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e, 0x56, 0x61, 0x6e,
	0x69, 0x6c, 0x6c, 0x61, 0xca, 0x02, 0x1f, 0x4e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x5c, 0x49,
	0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5c, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x5c, 0x56,
	0x61, 0x6e, 0x69, 0x6c, 0x6c, 0x61, 0xe2, 0x02, 0x2b, 0x4e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64,
	0x5c, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5c, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72,
	0x5c, 0x56, 0x61, 0x6e, 0x69, 0x6c, 0x6c, 0x61, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x22, 0x4e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x3a, 0x3a,
	0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x3a, 0x3a, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72,
	0x3a, 0x3a, 0x56, 0x61, 0x6e, 0x69, 0x6c, 0x6c, 0x61, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_pkg_drivers_instance_vanilla_driver_proto_rawDescOnce sync.Once
	file_pkg_drivers_instance_vanilla_driver_proto_rawDescData = file_pkg_drivers_instance_vanilla_driver_proto_rawDesc
)

func file_pkg_drivers_instance_vanilla_driver_proto_rawDescGZIP() []byte {
	file_pkg_drivers_instance_vanilla_driver_proto_rawDescOnce.Do(func() {
		file_pkg_drivers_instance_vanilla_driver_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_drivers_instance_vanilla_driver_proto_rawDescData)
	})
	return file_pkg_drivers_instance_vanilla_driver_proto_rawDescData
}

var file_pkg_drivers_instance_vanilla_driver_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_pkg_drivers_instance_vanilla_driver_proto_goTypes = []interface{}{
	(*GetTypeRequest)(nil),                         // 0: nocloud.instance.driver.vanilla.GetTypeRequest
	(*GetTypeResponse)(nil),                        // 1: nocloud.instance.driver.vanilla.GetTypeResponse
	(*UpRequest)(nil),                              // 2: nocloud.instance.driver.vanilla.UpRequest
	(*UpResponse)(nil),                             // 3: nocloud.instance.driver.vanilla.UpResponse
	(*DownRequest)(nil),                            // 4: nocloud.instance.driver.vanilla.DownRequest
	(*DownResponse)(nil),                           // 5: nocloud.instance.driver.vanilla.DownResponse
	(*ActionRequest)(nil),                          // 6: nocloud.instance.driver.vanilla.ActionRequest
	(*proto.InstancesGroup)(nil),                   // 7: nocloud.instances.InstancesGroup
	(*proto1.ServicesProvider)(nil),                // 8: nocloud.services_providers.ServicesProvider
	(*anypb.Any)(nil),                              // 9: google.protobuf.Any
	(*proto.TestInstancesGroupConfigRequest)(nil),  // 10: nocloud.instances.TestInstancesGroupConfigRequest
	(*proto1.TestResponse)(nil),                    // 11: nocloud.services_providers.TestResponse
	(*proto.TestInstancesGroupConfigResponse)(nil), // 12: nocloud.instances.TestInstancesGroupConfigResponse
}
var file_pkg_drivers_instance_vanilla_driver_proto_depIdxs = []int32{
	7,  // 0: nocloud.instance.driver.vanilla.UpRequest.group:type_name -> nocloud.instances.InstancesGroup
	8,  // 1: nocloud.instance.driver.vanilla.UpRequest.services_provider:type_name -> nocloud.services_providers.ServicesProvider
	7,  // 2: nocloud.instance.driver.vanilla.UpResponse.group:type_name -> nocloud.instances.InstancesGroup
	7,  // 3: nocloud.instance.driver.vanilla.DownRequest.group:type_name -> nocloud.instances.InstancesGroup
	8,  // 4: nocloud.instance.driver.vanilla.DownRequest.services_provider:type_name -> nocloud.services_providers.ServicesProvider
	8,  // 5: nocloud.instance.driver.vanilla.ActionRequest.services_provider:type_name -> nocloud.services_providers.ServicesProvider
	9,  // 6: nocloud.instance.driver.vanilla.ActionRequest.params:type_name -> google.protobuf.Any
	8,  // 7: nocloud.instance.driver.vanilla.DriverService.TestServiceProviderConfig:input_type -> nocloud.services_providers.ServicesProvider
	10, // 8: nocloud.instance.driver.vanilla.DriverService.TestInstancesGroupConfig:input_type -> nocloud.instances.TestInstancesGroupConfigRequest
	0,  // 9: nocloud.instance.driver.vanilla.DriverService.GetType:input_type -> nocloud.instance.driver.vanilla.GetTypeRequest
	2,  // 10: nocloud.instance.driver.vanilla.DriverService.Up:input_type -> nocloud.instance.driver.vanilla.UpRequest
	4,  // 11: nocloud.instance.driver.vanilla.DriverService.Down:input_type -> nocloud.instance.driver.vanilla.DownRequest
	6,  // 12: nocloud.instance.driver.vanilla.DriverService.Invoke:input_type -> nocloud.instance.driver.vanilla.ActionRequest
	11, // 13: nocloud.instance.driver.vanilla.DriverService.TestServiceProviderConfig:output_type -> nocloud.services_providers.TestResponse
	12, // 14: nocloud.instance.driver.vanilla.DriverService.TestInstancesGroupConfig:output_type -> nocloud.instances.TestInstancesGroupConfigResponse
	1,  // 15: nocloud.instance.driver.vanilla.DriverService.GetType:output_type -> nocloud.instance.driver.vanilla.GetTypeResponse
	3,  // 16: nocloud.instance.driver.vanilla.DriverService.Up:output_type -> nocloud.instance.driver.vanilla.UpResponse
	5,  // 17: nocloud.instance.driver.vanilla.DriverService.Down:output_type -> nocloud.instance.driver.vanilla.DownResponse
	9,  // 18: nocloud.instance.driver.vanilla.DriverService.Invoke:output_type -> google.protobuf.Any
	13, // [13:19] is the sub-list for method output_type
	7,  // [7:13] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_pkg_drivers_instance_vanilla_driver_proto_init() }
func file_pkg_drivers_instance_vanilla_driver_proto_init() {
	if File_pkg_drivers_instance_vanilla_driver_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTypeRequest); i {
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
		file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTypeResponse); i {
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
		file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpRequest); i {
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
		file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpResponse); i {
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
		file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownRequest); i {
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
		file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownResponse); i {
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
		file_pkg_drivers_instance_vanilla_driver_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ActionRequest); i {
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
			RawDescriptor: file_pkg_drivers_instance_vanilla_driver_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_drivers_instance_vanilla_driver_proto_goTypes,
		DependencyIndexes: file_pkg_drivers_instance_vanilla_driver_proto_depIdxs,
		MessageInfos:      file_pkg_drivers_instance_vanilla_driver_proto_msgTypes,
	}.Build()
	File_pkg_drivers_instance_vanilla_driver_proto = out.File
	file_pkg_drivers_instance_vanilla_driver_proto_rawDesc = nil
	file_pkg_drivers_instance_vanilla_driver_proto_goTypes = nil
	file_pkg_drivers_instance_vanilla_driver_proto_depIdxs = nil
}
