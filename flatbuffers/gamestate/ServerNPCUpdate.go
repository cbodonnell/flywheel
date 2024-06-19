// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package gamestate

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type ServerNPCUpdate struct {
	_tab flatbuffers.Table
}

func GetRootAsServerNPCUpdate(buf []byte, offset flatbuffers.UOffsetT) *ServerNPCUpdate {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &ServerNPCUpdate{}
	x.Init(buf, n+offset)
	return x
}

func FinishServerNPCUpdateBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.Finish(offset)
}

func GetSizePrefixedRootAsServerNPCUpdate(buf []byte, offset flatbuffers.UOffsetT) *ServerNPCUpdate {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &ServerNPCUpdate{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func FinishSizePrefixedServerNPCUpdateBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.FinishSizePrefixed(offset)
}

func (rcv *ServerNPCUpdate) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *ServerNPCUpdate) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *ServerNPCUpdate) Timestamp() int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *ServerNPCUpdate) MutateTimestamp(n int64) bool {
	return rcv._tab.MutateInt64Slot(4, n)
}

func (rcv *ServerNPCUpdate) NpcId() uint32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetUint32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *ServerNPCUpdate) MutateNpcId(n uint32) bool {
	return rcv._tab.MutateUint32Slot(6, n)
}

func (rcv *ServerNPCUpdate) NpcState(obj *NPCState) *NPCState {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
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

func ServerNPCUpdateStart(builder *flatbuffers.Builder) {
	builder.StartObject(3)
}
func ServerNPCUpdateAddTimestamp(builder *flatbuffers.Builder, timestamp int64) {
	builder.PrependInt64Slot(0, timestamp, 0)
}
func ServerNPCUpdateAddNpcId(builder *flatbuffers.Builder, npcId uint32) {
	builder.PrependUint32Slot(1, npcId, 0)
}
func ServerNPCUpdateAddNpcState(builder *flatbuffers.Builder, npcState flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(npcState), 0)
}
func ServerNPCUpdateEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}