// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package gamestate

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type NPCStateKeyValue struct {
	_tab flatbuffers.Table
}

func GetRootAsNPCStateKeyValue(buf []byte, offset flatbuffers.UOffsetT) *NPCStateKeyValue {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &NPCStateKeyValue{}
	x.Init(buf, n+offset)
	return x
}

func FinishNPCStateKeyValueBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.Finish(offset)
}

func GetSizePrefixedRootAsNPCStateKeyValue(buf []byte, offset flatbuffers.UOffsetT) *NPCStateKeyValue {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &NPCStateKeyValue{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func FinishSizePrefixedNPCStateKeyValueBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.FinishSizePrefixed(offset)
}

func (rcv *NPCStateKeyValue) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *NPCStateKeyValue) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *NPCStateKeyValue) Key() uint32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetUint32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *NPCStateKeyValue) MutateKey(n uint32) bool {
	return rcv._tab.MutateUint32Slot(4, n)
}

func (rcv *NPCStateKeyValue) Value(obj *NPCState) *NPCState {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(NPCState)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func NPCStateKeyValueStart(builder *flatbuffers.Builder) {
	builder.StartObject(2)
}
func NPCStateKeyValueAddKey(builder *flatbuffers.Builder, key uint32) {
	builder.PrependUint32Slot(0, key, 0)
}
func NPCStateKeyValueAddValue(builder *flatbuffers.Builder, value flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(value), 0)
}
func NPCStateKeyValueEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}