// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.25.0
// source: proto/structs.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ServiceState_State int32

const (
	ServiceState_Pending ServiceState_State = 0
	ServiceState_Running ServiceState_State = 1
	ServiceState_Tainted ServiceState_State = 2
	ServiceState_Dead    ServiceState_State = 3
)

// Enum value maps for ServiceState_State.
var (
	ServiceState_State_name = map[int32]string{
		0: "Pending",
		1: "Running",
		2: "Tainted",
		3: "Dead",
	}
	ServiceState_State_value = map[string]int32{
		"Pending": 0,
		"Running": 1,
		"Tainted": 2,
		"Dead":    3,
	}
)

func (x ServiceState_State) Enum() *ServiceState_State {
	p := new(ServiceState_State)
	*p = x
	return p
}

func (x ServiceState_State) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ServiceState_State) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_structs_proto_enumTypes[0].Descriptor()
}

func (ServiceState_State) Type() protoreflect.EnumType {
	return &file_proto_structs_proto_enumTypes[0]
}

func (x ServiceState_State) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ServiceState_State.Descriptor instead.
func (ServiceState_State) EnumDescriptor() ([]byte, []int) {
	return file_proto_structs_proto_rawDescGZIP(), []int{3, 0}
}

type Project struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name     string              `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Services map[string]*Service `protobuf:"bytes,2,rep,name=services,proto3" json:"services,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Project) Reset() {
	*x = Project{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_structs_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Project) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Project) ProtoMessage() {}

func (x *Project) ProtoReflect() protoreflect.Message {
	mi := &file_proto_structs_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Project.ProtoReflect.Descriptor instead.
func (*Project) Descriptor() ([]byte, []int) {
	return file_proto_structs_proto_rawDescGZIP(), []int{0}
}

func (x *Project) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Project) GetServices() map[string]*Service {
	if x != nil {
		return x.Services
	}
	return nil
}

type Service struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Image         string                 `protobuf:"bytes,1,opt,name=image,proto3" json:"image,omitempty"`
	Depends       []string               `protobuf:"bytes,2,rep,name=depends,proto3" json:"depends,omitempty"`
	Env           map[string]string      `protobuf:"bytes,3,rep,name=env,proto3" json:"env,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Args          []string               `protobuf:"bytes,4,rep,name=args,proto3" json:"args,omitempty"`
	Mounts        []*Service_MountConfig `protobuf:"bytes,5,rep,name=mounts,proto3" json:"mounts,omitempty"`
	NetworkMode   string                 `protobuf:"bytes,6,opt,name=networkMode,proto3" json:"networkMode,omitempty"`
	RestartPolicy *Service_RestartPolicy `protobuf:"bytes,7,opt,name=restartPolicy,proto3" json:"restartPolicy,omitempty"`
}

func (x *Service) Reset() {
	*x = Service{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_structs_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Service) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Service) ProtoMessage() {}

func (x *Service) ProtoReflect() protoreflect.Message {
	mi := &file_proto_structs_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Service.ProtoReflect.Descriptor instead.
func (*Service) Descriptor() ([]byte, []int) {
	return file_proto_structs_proto_rawDescGZIP(), []int{1}
}

func (x *Service) GetImage() string {
	if x != nil {
		return x.Image
	}
	return ""
}

func (x *Service) GetDepends() []string {
	if x != nil {
		return x.Depends
	}
	return nil
}

func (x *Service) GetEnv() map[string]string {
	if x != nil {
		return x.Env
	}
	return nil
}

func (x *Service) GetArgs() []string {
	if x != nil {
		return x.Args
	}
	return nil
}

func (x *Service) GetMounts() []*Service_MountConfig {
	if x != nil {
		return x.Mounts
	}
	return nil
}

func (x *Service) GetNetworkMode() string {
	if x != nil {
		return x.NetworkMode
	}
	return ""
}

func (x *Service) GetRestartPolicy() *Service_RestartPolicy {
	if x != nil {
		return x.RestartPolicy
	}
	return nil
}

type ExitResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ExitCode uint64 `protobuf:"varint,1,opt,name=exitCode,proto3" json:"exitCode,omitempty"`
}

func (x *ExitResult) Reset() {
	*x = ExitResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_structs_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExitResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExitResult) ProtoMessage() {}

func (x *ExitResult) ProtoReflect() protoreflect.Message {
	mi := &file_proto_structs_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExitResult.ProtoReflect.Descriptor instead.
func (*ExitResult) Descriptor() ([]byte, []int) {
	return file_proto_structs_proto_rawDescGZIP(), []int{2}
}

func (x *ExitResult) GetExitCode() uint64 {
	if x != nil {
		return x.ExitCode
	}
	return 0
}

type ServiceState struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	State  ServiceState_State   `protobuf:"varint,1,opt,name=state,proto3,enum=proto.ServiceState_State" json:"state,omitempty"`
	Handle *ServiceState_Handle `protobuf:"bytes,2,opt,name=handle,proto3" json:"handle,omitempty"`
	Hash   string               `protobuf:"bytes,3,opt,name=hash,proto3" json:"hash,omitempty"`
	Failed bool                 `protobuf:"varint,4,opt,name=failed,proto3" json:"failed,omitempty"`
}

func (x *ServiceState) Reset() {
	*x = ServiceState{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_structs_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceState) ProtoMessage() {}

func (x *ServiceState) ProtoReflect() protoreflect.Message {
	mi := &file_proto_structs_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceState.ProtoReflect.Descriptor instead.
func (*ServiceState) Descriptor() ([]byte, []int) {
	return file_proto_structs_proto_rawDescGZIP(), []int{3}
}

func (x *ServiceState) GetState() ServiceState_State {
	if x != nil {
		return x.State
	}
	return ServiceState_Pending
}

func (x *ServiceState) GetHandle() *ServiceState_Handle {
	if x != nil {
		return x.Handle
	}
	return nil
}

func (x *ServiceState) GetHash() string {
	if x != nil {
		return x.Hash
	}
	return ""
}

func (x *ServiceState) GetFailed() bool {
	if x != nil {
		return x.Failed
	}
	return false
}

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Project string                 `protobuf:"bytes,1,opt,name=project,proto3" json:"project,omitempty"`
	Service string                 `protobuf:"bytes,2,opt,name=service,proto3" json:"service,omitempty"`
	Type    string                 `protobuf:"bytes,3,opt,name=type,proto3" json:"type,omitempty"`
	Details map[string]string      `protobuf:"bytes,4,rep,name=details,proto3" json:"details,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Time    *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=time,proto3" json:"time,omitempty"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_structs_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_proto_structs_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_proto_structs_proto_rawDescGZIP(), []int{4}
}

func (x *Event) GetProject() string {
	if x != nil {
		return x.Project
	}
	return ""
}

func (x *Event) GetService() string {
	if x != nil {
		return x.Service
	}
	return ""
}

func (x *Event) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Event) GetDetails() map[string]string {
	if x != nil {
		return x.Details
	}
	return nil
}

func (x *Event) GetTime() *timestamppb.Timestamp {
	if x != nil {
		return x.Time
	}
	return nil
}

type ExecTaskResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Stdout   string `protobuf:"bytes,1,opt,name=stdout,proto3" json:"stdout,omitempty"`
	Stderr   string `protobuf:"bytes,2,opt,name=stderr,proto3" json:"stderr,omitempty"`
	ExitCode uint64 `protobuf:"varint,3,opt,name=exitCode,proto3" json:"exitCode,omitempty"`
}

func (x *ExecTaskResult) Reset() {
	*x = ExecTaskResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_structs_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecTaskResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecTaskResult) ProtoMessage() {}

func (x *ExecTaskResult) ProtoReflect() protoreflect.Message {
	mi := &file_proto_structs_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecTaskResult.ProtoReflect.Descriptor instead.
func (*ExecTaskResult) Descriptor() ([]byte, []int) {
	return file_proto_structs_proto_rawDescGZIP(), []int{5}
}

func (x *ExecTaskResult) GetStdout() string {
	if x != nil {
		return x.Stdout
	}
	return ""
}

func (x *ExecTaskResult) GetStderr() string {
	if x != nil {
		return x.Stderr
	}
	return ""
}

func (x *ExecTaskResult) GetExitCode() uint64 {
	if x != nil {
		return x.ExitCode
	}
	return 0
}

type Project_Ref struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *Project_Ref) Reset() {
	*x = Project_Ref{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_structs_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Project_Ref) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Project_Ref) ProtoMessage() {}

func (x *Project_Ref) ProtoReflect() protoreflect.Message {
	mi := &file_proto_structs_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Project_Ref.ProtoReflect.Descriptor instead.
func (*Project_Ref) Descriptor() ([]byte, []int) {
	return file_proto_structs_proto_rawDescGZIP(), []int{0, 1}
}

func (x *Project_Ref) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type Service_MountConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HostPath string `protobuf:"bytes,1,opt,name=hostPath,proto3" json:"hostPath,omitempty"`
	TaskPath string `protobuf:"bytes,2,opt,name=taskPath,proto3" json:"taskPath,omitempty"`
}

func (x *Service_MountConfig) Reset() {
	*x = Service_MountConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_structs_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Service_MountConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Service_MountConfig) ProtoMessage() {}

func (x *Service_MountConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proto_structs_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Service_MountConfig.ProtoReflect.Descriptor instead.
func (*Service_MountConfig) Descriptor() ([]byte, []int) {
	return file_proto_structs_proto_rawDescGZIP(), []int{1, 1}
}

func (x *Service_MountConfig) GetHostPath() string {
	if x != nil {
		return x.HostPath
	}
	return ""
}

func (x *Service_MountConfig) GetTaskPath() string {
	if x != nil {
		return x.TaskPath
	}
	return ""
}

type Service_RestartPolicy struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Service_RestartPolicy) Reset() {
	*x = Service_RestartPolicy{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_structs_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Service_RestartPolicy) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Service_RestartPolicy) ProtoMessage() {}

func (x *Service_RestartPolicy) ProtoReflect() protoreflect.Message {
	mi := &file_proto_structs_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Service_RestartPolicy.ProtoReflect.Descriptor instead.
func (*Service_RestartPolicy) Descriptor() ([]byte, []int) {
	return file_proto_structs_proto_rawDescGZIP(), []int{1, 2}
}

type ServiceState_Handle struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ContainerId string `protobuf:"bytes,1,opt,name=containerId,proto3" json:"containerId,omitempty"`
	Ip          string `protobuf:"bytes,2,opt,name=ip,proto3" json:"ip,omitempty"`
}

func (x *ServiceState_Handle) Reset() {
	*x = ServiceState_Handle{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_structs_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceState_Handle) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceState_Handle) ProtoMessage() {}

func (x *ServiceState_Handle) ProtoReflect() protoreflect.Message {
	mi := &file_proto_structs_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceState_Handle.ProtoReflect.Descriptor instead.
func (*ServiceState_Handle) Descriptor() ([]byte, []int) {
	return file_proto_structs_proto_rawDescGZIP(), []int{3, 0}
}

func (x *ServiceState_Handle) GetContainerId() string {
	if x != nil {
		return x.ContainerId
	}
	return ""
}

func (x *ServiceState_Handle) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

var File_proto_structs_proto protoreflect.FileDescriptor

var file_proto_structs_proto_rawDesc = []byte{
	0x0a, 0x13, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbb, 0x01,
	0x0a, 0x07, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x38, 0x0a,
	0x08, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2e,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x1a, 0x4b, 0x0a, 0x0d, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x24, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x1a, 0x15, 0x0a, 0x03, 0x52, 0x65, 0x66, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0xa2, 0x03, 0x0a, 0x07,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x69, 0x6d, 0x61, 0x67, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x12, 0x18, 0x0a,
	0x07, 0x64, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07,
	0x64, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x73, 0x12, 0x29, 0x0a, 0x03, 0x65, 0x6e, 0x76, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x45, 0x6e, 0x76, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x03, 0x65,
	0x6e, 0x76, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x72, 0x67, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x04, 0x61, 0x72, 0x67, 0x73, 0x12, 0x32, 0x0a, 0x06, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x73,
	0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x4d, 0x6f, 0x75, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x52, 0x06, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x73, 0x12, 0x20, 0x0a, 0x0b, 0x6e, 0x65,
	0x74, 0x77, 0x6f, 0x72, 0x6b, 0x4d, 0x6f, 0x64, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0b, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x4d, 0x6f, 0x64, 0x65, 0x12, 0x42, 0x0a, 0x0d,
	0x72, 0x65, 0x73, 0x74, 0x61, 0x72, 0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x2e, 0x52, 0x65, 0x73, 0x74, 0x61, 0x72, 0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63,
	0x79, 0x52, 0x0d, 0x72, 0x65, 0x73, 0x74, 0x61, 0x72, 0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79,
	0x1a, 0x36, 0x0a, 0x08, 0x45, 0x6e, 0x76, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x45, 0x0a, 0x0b, 0x4d, 0x6f, 0x75, 0x6e,
	0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x50,
	0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x50,
	0x61, 0x74, 0x68, 0x12, 0x1a, 0x0a, 0x08, 0x74, 0x61, 0x73, 0x6b, 0x50, 0x61, 0x74, 0x68, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x61, 0x73, 0x6b, 0x50, 0x61, 0x74, 0x68, 0x1a,
	0x0f, 0x0a, 0x0d, 0x52, 0x65, 0x73, 0x74, 0x61, 0x72, 0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79,
	0x22, 0x28, 0x0a, 0x0a, 0x45, 0x78, 0x69, 0x74, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x1a,
	0x0a, 0x08, 0x65, 0x78, 0x69, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x08, 0x65, 0x78, 0x69, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x22, 0x95, 0x02, 0x0a, 0x0c, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x2f, 0x0a, 0x05, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x19, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x53, 0x74, 0x61, 0x74, 0x65, 0x2e,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x32, 0x0a, 0x06,
	0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x2e, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x52, 0x06, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x68, 0x61, 0x73, 0x68, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x1a, 0x3a, 0x0a, 0x06,
	0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69,
	0x6e, 0x65, 0x72, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x6e,
	0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x49, 0x64, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x70, 0x22, 0x38, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x12, 0x0b, 0x0a, 0x07, 0x50, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x10, 0x00, 0x12, 0x0b,
	0x0a, 0x07, 0x52, 0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x54,
	0x61, 0x69, 0x6e, 0x74, 0x65, 0x64, 0x10, 0x02, 0x12, 0x08, 0x0a, 0x04, 0x44, 0x65, 0x61, 0x64,
	0x10, 0x03, 0x22, 0xf0, 0x01, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07,
	0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x12, 0x33, 0x0a, 0x07, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x18,
	0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x2e, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x07, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x12, 0x2e, 0x0a, 0x04, 0x74, 0x69, 0x6d,
	0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x1a, 0x3a, 0x0a, 0x0c, 0x44, 0x65, 0x74,
	0x61, 0x69, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x5c, 0x0a, 0x0e, 0x45, 0x78, 0x65, 0x63, 0x54, 0x61, 0x73,
	0x6b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x64, 0x6f, 0x75,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x64, 0x6f, 0x75, 0x74, 0x12,
	0x16, 0x0a, 0x06, 0x73, 0x74, 0x64, 0x65, 0x72, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x73, 0x74, 0x64, 0x65, 0x72, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x78, 0x69, 0x74, 0x43,
	0x6f, 0x64, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x65, 0x78, 0x69, 0x74, 0x43,
	0x6f, 0x64, 0x65, 0x42, 0x08, 0x5a, 0x06, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_structs_proto_rawDescOnce sync.Once
	file_proto_structs_proto_rawDescData = file_proto_structs_proto_rawDesc
)

func file_proto_structs_proto_rawDescGZIP() []byte {
	file_proto_structs_proto_rawDescOnce.Do(func() {
		file_proto_structs_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_structs_proto_rawDescData)
	})
	return file_proto_structs_proto_rawDescData
}

var file_proto_structs_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_structs_proto_msgTypes = make([]protoimpl.MessageInfo, 13)
var file_proto_structs_proto_goTypes = []interface{}{
	(ServiceState_State)(0),       // 0: proto.ServiceState.State
	(*Project)(nil),               // 1: proto.Project
	(*Service)(nil),               // 2: proto.Service
	(*ExitResult)(nil),            // 3: proto.ExitResult
	(*ServiceState)(nil),          // 4: proto.ServiceState
	(*Event)(nil),                 // 5: proto.Event
	(*ExecTaskResult)(nil),        // 6: proto.ExecTaskResult
	nil,                           // 7: proto.Project.ServicesEntry
	(*Project_Ref)(nil),           // 8: proto.Project.Ref
	nil,                           // 9: proto.Service.EnvEntry
	(*Service_MountConfig)(nil),   // 10: proto.Service.MountConfig
	(*Service_RestartPolicy)(nil), // 11: proto.Service.RestartPolicy
	(*ServiceState_Handle)(nil),   // 12: proto.ServiceState.Handle
	nil,                           // 13: proto.Event.DetailsEntry
	(*timestamppb.Timestamp)(nil), // 14: google.protobuf.Timestamp
}
var file_proto_structs_proto_depIdxs = []int32{
	7,  // 0: proto.Project.services:type_name -> proto.Project.ServicesEntry
	9,  // 1: proto.Service.env:type_name -> proto.Service.EnvEntry
	10, // 2: proto.Service.mounts:type_name -> proto.Service.MountConfig
	11, // 3: proto.Service.restartPolicy:type_name -> proto.Service.RestartPolicy
	0,  // 4: proto.ServiceState.state:type_name -> proto.ServiceState.State
	12, // 5: proto.ServiceState.handle:type_name -> proto.ServiceState.Handle
	13, // 6: proto.Event.details:type_name -> proto.Event.DetailsEntry
	14, // 7: proto.Event.time:type_name -> google.protobuf.Timestamp
	2,  // 8: proto.Project.ServicesEntry.value:type_name -> proto.Service
	9,  // [9:9] is the sub-list for method output_type
	9,  // [9:9] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_proto_structs_proto_init() }
func file_proto_structs_proto_init() {
	if File_proto_structs_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_structs_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Project); i {
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
		file_proto_structs_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Service); i {
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
		file_proto_structs_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExitResult); i {
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
		file_proto_structs_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServiceState); i {
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
		file_proto_structs_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
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
		file_proto_structs_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecTaskResult); i {
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
		file_proto_structs_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Project_Ref); i {
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
		file_proto_structs_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Service_MountConfig); i {
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
		file_proto_structs_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Service_RestartPolicy); i {
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
		file_proto_structs_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServiceState_Handle); i {
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
			RawDescriptor: file_proto_structs_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   13,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_structs_proto_goTypes,
		DependencyIndexes: file_proto_structs_proto_depIdxs,
		EnumInfos:         file_proto_structs_proto_enumTypes,
		MessageInfos:      file_proto_structs_proto_msgTypes,
	}.Build()
	File_proto_structs_proto = out.File
	file_proto_structs_proto_rawDesc = nil
	file_proto_structs_proto_goTypes = nil
	file_proto_structs_proto_depIdxs = nil
}
