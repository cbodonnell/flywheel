// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package gamestate

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type NPCState struct {
	_tab flatbuffers.Table
}

func GetRootAsNPCState(buf []byte, offset flatbuffers.UOffsetT) *NPCState {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &NPCState{}
	x.Init(buf, n+offset)
	return x
}

func FinishNPCStateBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.Finish(offset)
}

func GetSizePrefixedRootAsNPCState(buf []byte, offset flatbuffers.UOffsetT) *NPCState {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &NPCState{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func FinishSizePrefixedNPCStateBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.FinishSizePrefixed(offset)
}

func (rcv *NPCState) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *NPCState) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *NPCState) Position(obj *Position) *Position {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(Position)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *NPCState) Velocity(obj *Velocity) *Velocity {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(Velocity)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *NPCState) IsOnGround() bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetBool(o + rcv._tab.Pos)
	}
	return false
}

func (rcv *NPCState) MutateIsOnGround(n bool) bool {
	return rcv._tab.MutateBoolSlot(8, n)
}

func (rcv *NPCState) Animation() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *NPCState) MutateAnimation(n byte) bool {
	return rcv._tab.MutateByteSlot(10, n)
}

func (rcv *NPCState) FlipH() bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.GetBool(o + rcv._tab.Pos)
	}
	return false
}

func (rcv *NPCState) MutateFlipH(n bool) bool {
	return rcv._tab.MutateBoolSlot(12, n)
}

func (rcv *NPCState) AnimationSequence() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *NPCState) MutateAnimationSequence(n byte) bool {
	return rcv._tab.MutateByteSlot(14, n)
}

func (rcv *NPCState) Hitpoints() int16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *NPCState) MutateHitpoints(n int16) bool {
	return rcv._tab.MutateInt16Slot(16, n)
}

func NPCStateStart(builder *flatbuffers.Builder) {
	builder.StartObject(7)
}
func NPCStateAddPosition(builder *flatbuffers.Builder, position flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(position), 0)
}
func NPCStateAddVelocity(builder *flatbuffers.Builder, velocity flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(velocity), 0)
}
func NPCStateAddIsOnGround(builder *flatbuffers.Builder, isOnGround bool) {
	builder.PrependBoolSlot(2, isOnGround, false)
}
func NPCStateAddAnimation(builder *flatbuffers.Builder, animation byte) {
	builder.PrependByteSlot(3, animation, 0)
}
func NPCStateAddFlipH(builder *flatbuffers.Builder, flipH bool) {
	builder.PrependBoolSlot(4, flipH, false)
}
func NPCStateAddAnimationSequence(builder *flatbuffers.Builder, animationSequence byte) {
	builder.PrependByteSlot(5, animationSequence, 0)
}
func NPCStateAddHitpoints(builder *flatbuffers.Builder, hitpoints int16) {
	builder.PrependInt16Slot(6, hitpoints, 0)
}
func NPCStateEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
