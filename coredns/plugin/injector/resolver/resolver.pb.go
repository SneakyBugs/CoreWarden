// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: resolver.proto

package resolver

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Question struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name  string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Qtype uint32 `protobuf:"varint,2,opt,name=qtype,proto3" json:"qtype,omitempty"` // No need for qclass since it is always INET.
}

func (x *Question) Reset() {
	*x = Question{}
	if protoimpl.UnsafeEnabled {
		mi := &file_resolver_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Question) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Question) ProtoMessage() {}

func (x *Question) ProtoReflect() protoreflect.Message {
	mi := &file_resolver_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Question.ProtoReflect.Descriptor instead.
func (*Question) Descriptor() ([]byte, []int) {
	return file_resolver_proto_rawDescGZIP(), []int{0}
}

func (x *Question) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Question) GetQtype() uint32 {
	if x != nil {
		return x.Qtype
	}
	return 0
}

type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Answer []string `protobuf:"bytes,1,rep,name=answer,proto3" json:"answer,omitempty"`
	Ns     []string `protobuf:"bytes,2,rep,name=ns,proto3" json:"ns,omitempty"`
	Extra  []string `protobuf:"bytes,3,rep,name=extra,proto3" json:"extra,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_resolver_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_resolver_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_resolver_proto_rawDescGZIP(), []int{1}
}

func (x *Response) GetAnswer() []string {
	if x != nil {
		return x.Answer
	}
	return nil
}

func (x *Response) GetNs() []string {
	if x != nil {
		return x.Ns
	}
	return nil
}

func (x *Response) GetExtra() []string {
	if x != nil {
		return x.Extra
	}
	return nil
}

var File_resolver_proto protoreflect.FileDescriptor

var file_resolver_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x72, 0x65, 0x73, 0x6f, 0x6c, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x6c, 0x76, 0x65, 0x72, 0x22, 0x34, 0x0a, 0x08, 0x51, 0x75,
	0x65, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x71, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x71, 0x74, 0x79, 0x70, 0x65,
	0x22, 0x48, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06,
	0x61, 0x6e, 0x73, 0x77, 0x65, 0x72, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x61, 0x6e,
	0x73, 0x77, 0x65, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x02, 0x6e, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x78, 0x74, 0x72, 0x61, 0x18, 0x03, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x05, 0x65, 0x78, 0x74, 0x72, 0x61, 0x32, 0x3f, 0x0a, 0x08, 0x52, 0x65,
	0x73, 0x6f, 0x6c, 0x76, 0x65, 0x72, 0x12, 0x33, 0x0a, 0x07, 0x52, 0x65, 0x73, 0x6f, 0x6c, 0x76,
	0x65, 0x12, 0x12, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x6c, 0x76, 0x65, 0x72, 0x2e, 0x51, 0x75, 0x65,
	0x73, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x12, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x6c, 0x76, 0x65, 0x72,
	0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x32, 0x5a, 0x30, 0x67,
	0x69, 0x74, 0x2e, 0x68, 0x6f, 0x75, 0x73, 0x65, 0x6f, 0x66, 0x6b, 0x75, 0x6d, 0x6d, 0x65, 0x72,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x6f, 0x72, 0x2f, 0x68, 0x6f, 0x6d, 0x65, 0x2d, 0x64,
	0x6e, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x6c, 0x76, 0x65, 0x72, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_resolver_proto_rawDescOnce sync.Once
	file_resolver_proto_rawDescData = file_resolver_proto_rawDesc
)

func file_resolver_proto_rawDescGZIP() []byte {
	file_resolver_proto_rawDescOnce.Do(func() {
		file_resolver_proto_rawDescData = protoimpl.X.CompressGZIP(file_resolver_proto_rawDescData)
	})
	return file_resolver_proto_rawDescData
}

var file_resolver_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_resolver_proto_goTypes = []interface{}{
	(*Question)(nil), // 0: resolver.Question
	(*Response)(nil), // 1: resolver.Response
}
var file_resolver_proto_depIdxs = []int32{
	0, // 0: resolver.Resolver.Resolve:input_type -> resolver.Question
	1, // 1: resolver.Resolver.Resolve:output_type -> resolver.Response
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_resolver_proto_init() }
func file_resolver_proto_init() {
	if File_resolver_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_resolver_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Question); i {
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
		file_resolver_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
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
			RawDescriptor: file_resolver_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_resolver_proto_goTypes,
		DependencyIndexes: file_resolver_proto_depIdxs,
		MessageInfos:      file_resolver_proto_msgTypes,
	}.Build()
	File_resolver_proto = out.File
	file_resolver_proto_rawDesc = nil
	file_resolver_proto_goTypes = nil
	file_resolver_proto_depIdxs = nil
}
