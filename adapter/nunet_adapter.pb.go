// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.12.4
// source: nunet_adapter.proto

package adapter

import (
	_ "github.com/golang/protobuf/ptypes/empty"
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

type GetPeerParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetPeerParams) Reset() {
	*x = GetPeerParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nunet_adapter_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPeerParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPeerParams) ProtoMessage() {}

func (x *GetPeerParams) ProtoReflect() protoreflect.Message {
	mi := &file_nunet_adapter_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPeerParams.ProtoReflect.Descriptor instead.
func (*GetPeerParams) Descriptor() ([]byte, []int) {
	return file_nunet_adapter_proto_rawDescGZIP(), []int{0}
}

type GetDhtParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetDhtParams) Reset() {
	*x = GetDhtParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nunet_adapter_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDhtParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDhtParams) ProtoMessage() {}

func (x *GetDhtParams) ProtoReflect() protoreflect.Message {
	mi := &file_nunet_adapter_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDhtParams.ProtoReflect.Descriptor instead.
func (*GetDhtParams) Descriptor() ([]byte, []int) {
	return file_nunet_adapter_proto_rawDescGZIP(), []int{1}
}

type Peers struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PeerAddress string `protobuf:"bytes,1,opt,name=peer_address,json=peerAddress,proto3" json:"peer_address,omitempty"`
}

func (x *Peers) Reset() {
	*x = Peers{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nunet_adapter_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Peers) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Peers) ProtoMessage() {}

func (x *Peers) ProtoReflect() protoreflect.Message {
	mi := &file_nunet_adapter_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Peers.ProtoReflect.Descriptor instead.
func (*Peers) Descriptor() ([]byte, []int) {
	return file_nunet_adapter_proto_rawDescGZIP(), []int{2}
}

func (x *Peers) GetPeerAddress() string {
	if x != nil {
		return x.PeerAddress
	}
	return ""
}

type DhtContents struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DhtContents string `protobuf:"bytes,1,opt,name=dht_contents,json=dhtContents,proto3" json:"dht_contents,omitempty"`
}

func (x *DhtContents) Reset() {
	*x = DhtContents{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nunet_adapter_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DhtContents) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DhtContents) ProtoMessage() {}

func (x *DhtContents) ProtoReflect() protoreflect.Message {
	mi := &file_nunet_adapter_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DhtContents.ProtoReflect.Descriptor instead.
func (*DhtContents) Descriptor() ([]byte, []int) {
	return file_nunet_adapter_proto_rawDescGZIP(), []int{3}
}

func (x *DhtContents) GetDhtContents() string {
	if x != nil {
		return x.DhtContents
	}
	return ""
}

type MessageParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PeerId         string `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	MessageContent string `protobuf:"bytes,2,opt,name=message_content,json=messageContent,proto3" json:"message_content,omitempty"`
}

func (x *MessageParams) Reset() {
	*x = MessageParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nunet_adapter_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageParams) ProtoMessage() {}

func (x *MessageParams) ProtoReflect() protoreflect.Message {
	mi := &file_nunet_adapter_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageParams.ProtoReflect.Descriptor instead.
func (*MessageParams) Descriptor() ([]byte, []int) {
	return file_nunet_adapter_proto_rawDescGZIP(), []int{4}
}

func (x *MessageParams) GetPeerId() string {
	if x != nil {
		return x.PeerId
	}
	return ""
}

func (x *MessageParams) GetMessageContent() string {
	if x != nil {
		return x.MessageContent
	}
	return ""
}

type MessageResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MessageResponse string `protobuf:"bytes,1,opt,name=message_response,json=messageResponse,proto3" json:"message_response,omitempty"`
}

func (x *MessageResponse) Reset() {
	*x = MessageResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nunet_adapter_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageResponse) ProtoMessage() {}

func (x *MessageResponse) ProtoReflect() protoreflect.Message {
	mi := &file_nunet_adapter_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageResponse.ProtoReflect.Descriptor instead.
func (*MessageResponse) Descriptor() ([]byte, []int) {
	return file_nunet_adapter_proto_rawDescGZIP(), []int{5}
}

func (x *MessageResponse) GetMessageResponse() string {
	if x != nil {
		return x.MessageResponse
	}
	return ""
}

type ReceivedMessageParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ReceivedMessageParams) Reset() {
	*x = ReceivedMessageParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nunet_adapter_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReceivedMessageParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReceivedMessageParams) ProtoMessage() {}

func (x *ReceivedMessageParams) ProtoReflect() protoreflect.Message {
	mi := &file_nunet_adapter_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReceivedMessageParams.ProtoReflect.Descriptor instead.
func (*ReceivedMessageParams) Descriptor() ([]byte, []int) {
	return file_nunet_adapter_proto_rawDescGZIP(), []int{6}
}

var File_nunet_adapter_proto protoreflect.FileDescriptor

var file_nunet_adapter_proto_rawDesc = []byte{
	0x0a, 0x13, 0x6e, 0x75, 0x6e, 0x65, 0x74, 0x5f, 0x61, 0x64, 0x61, 0x70, 0x74, 0x65, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x0f, 0x0a, 0x0d, 0x67, 0x65, 0x74, 0x50, 0x65, 0x65, 0x72, 0x50, 0x61, 0x72,
	0x61, 0x6d, 0x73, 0x22, 0x0e, 0x0a, 0x0c, 0x67, 0x65, 0x74, 0x44, 0x68, 0x74, 0x50, 0x61, 0x72,
	0x61, 0x6d, 0x73, 0x22, 0x2a, 0x0a, 0x05, 0x70, 0x65, 0x65, 0x72, 0x73, 0x12, 0x21, 0x0a, 0x0c,
	0x70, 0x65, 0x65, 0x72, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0b, 0x70, 0x65, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x22,
	0x30, 0x0a, 0x0b, 0x44, 0x68, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x21,
	0x0a, 0x0c, 0x64, 0x68, 0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x68, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x73, 0x22, 0x51, 0x0a, 0x0d, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x50, 0x61, 0x72, 0x61,
	0x6d, 0x73, 0x12, 0x17, 0x0a, 0x07, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x65, 0x65, 0x72, 0x49, 0x64, 0x12, 0x27, 0x0a, 0x0f, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x22, 0x3c, 0x0a, 0x0f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x29, 0x0a, 0x10, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x17, 0x0a, 0x15, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x64, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x32, 0xd9, 0x01, 0x0a, 0x0c,
	0x4e, 0x75, 0x6e, 0x65, 0x74, 0x41, 0x64, 0x61, 0x70, 0x74, 0x65, 0x72, 0x12, 0x27, 0x0a, 0x0b,
	0x67, 0x65, 0x74, 0x50, 0x65, 0x65, 0x72, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x0e, 0x2e, 0x67, 0x65,
	0x74, 0x50, 0x65, 0x65, 0x72, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x1a, 0x06, 0x2e, 0x70, 0x65,
	0x65, 0x72, 0x73, 0x22, 0x00, 0x12, 0x2e, 0x0a, 0x0d, 0x67, 0x65, 0x74, 0x44, 0x68, 0x74, 0x43,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x0d, 0x2e, 0x67, 0x65, 0x74, 0x44, 0x68, 0x74, 0x50,
	0x61, 0x72, 0x61, 0x6d, 0x73, 0x1a, 0x0c, 0x2e, 0x44, 0x68, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0x22, 0x00, 0x12, 0x31, 0x0a, 0x0b, 0x73, 0x65, 0x6e, 0x64, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x0e, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x50, 0x61,
	0x72, 0x61, 0x6d, 0x73, 0x1a, 0x10, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x3d, 0x0a, 0x0f, 0x72, 0x65, 0x63, 0x65,
	0x69, 0x76, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x16, 0x2e, 0x72, 0x65,
	0x63, 0x65, 0x69, 0x76, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x50, 0x61, 0x72,
	0x61, 0x6d, 0x73, 0x1a, 0x10, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x34, 0x5a, 0x32, 0x67, 0x69, 0x74, 0x6c, 0x61,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6e, 0x75, 0x6e, 0x65, 0x74, 0x2f, 0x64, 0x65, 0x76, 0x69,
	0x63, 0x65, 0x2d, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x2d, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x61, 0x64, 0x61, 0x70, 0x74, 0x65, 0x72, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_nunet_adapter_proto_rawDescOnce sync.Once
	file_nunet_adapter_proto_rawDescData = file_nunet_adapter_proto_rawDesc
)

func file_nunet_adapter_proto_rawDescGZIP() []byte {
	file_nunet_adapter_proto_rawDescOnce.Do(func() {
		file_nunet_adapter_proto_rawDescData = protoimpl.X.CompressGZIP(file_nunet_adapter_proto_rawDescData)
	})
	return file_nunet_adapter_proto_rawDescData
}

var file_nunet_adapter_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_nunet_adapter_proto_goTypes = []interface{}{
	(*GetPeerParams)(nil),         // 0: getPeerParams
	(*GetDhtParams)(nil),          // 1: getDhtParams
	(*Peers)(nil),                 // 2: peers
	(*DhtContents)(nil),           // 3: DhtContents
	(*MessageParams)(nil),         // 4: messageParams
	(*MessageResponse)(nil),       // 5: messageResponse
	(*ReceivedMessageParams)(nil), // 6: receivedMessageParams
}
var file_nunet_adapter_proto_depIdxs = []int32{
	0, // 0: NunetAdapter.getPeerList:input_type -> getPeerParams
	1, // 1: NunetAdapter.getDhtContent:input_type -> getDhtParams
	4, // 2: NunetAdapter.sendMessage:input_type -> messageParams
	6, // 3: NunetAdapter.receivedMessage:input_type -> receivedMessageParams
	2, // 4: NunetAdapter.getPeerList:output_type -> peers
	3, // 5: NunetAdapter.getDhtContent:output_type -> DhtContents
	5, // 6: NunetAdapter.sendMessage:output_type -> messageResponse
	5, // 7: NunetAdapter.receivedMessage:output_type -> messageResponse
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_nunet_adapter_proto_init() }
func file_nunet_adapter_proto_init() {
	if File_nunet_adapter_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_nunet_adapter_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPeerParams); i {
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
		file_nunet_adapter_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDhtParams); i {
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
		file_nunet_adapter_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Peers); i {
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
		file_nunet_adapter_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DhtContents); i {
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
		file_nunet_adapter_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageParams); i {
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
		file_nunet_adapter_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageResponse); i {
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
		file_nunet_adapter_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReceivedMessageParams); i {
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
			RawDescriptor: file_nunet_adapter_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_nunet_adapter_proto_goTypes,
		DependencyIndexes: file_nunet_adapter_proto_depIdxs,
		MessageInfos:      file_nunet_adapter_proto_msgTypes,
	}.Build()
	File_nunet_adapter_proto = out.File
	file_nunet_adapter_proto_rawDesc = nil
	file_nunet_adapter_proto_goTypes = nil
	file_nunet_adapter_proto_depIdxs = nil
}
