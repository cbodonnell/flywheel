// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package gamestate

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type Velocity struct {
	_tab flatbuffers.Table
}

func GetRootAsVelocity(buf []byte, offset flatbuffers.UOffsetT) *Velocity {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Velocity{}
	x.Init(buf, n+offset)
	return x
}

func FinishVelocityBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.Finish(offset)
}

func GetSizePrefixedRootAsVelocity(buf []byte, offset flatbuffers.UOffsetT) *Velocity {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &Velocity{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func FinishSizePrefixedVelocityBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	builder.FinishSizePrefixed(offset)
}

func (rcv *Velocity) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Velocity) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Velocity) X() float64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetFloat64(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Velocity) MutateX(n float64) bool {
	return rcv._tab.MutateFloat64Slot(4, n)
}

func (rcv *Velocity) Y() float64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetFloat64(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Velocity) MutateY(n float64) bool {
	return rcv._tab.MutateFloat64Slot(6, n)
}

func VelocityStart(builder *flatbuffers.Builder) {
	builder.StartObject(2)
}
func VelocityAddX(builder *flatbuffers.Builder, x float64) {
	builder.PrependFloat64Slot(0, x, 0.0)
}
func VelocityAddY(builder *flatbuffers.Builder, y float64) {
	builder.PrependFloat64Slot(1, y, 0.0)
}
func VelocityEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
