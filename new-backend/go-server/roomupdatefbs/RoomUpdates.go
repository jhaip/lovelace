// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package roomupdatefbs

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type RoomUpdates struct {
	_tab flatbuffers.Table
}

func GetRootAsRoomUpdates(buf []byte, offset flatbuffers.UOffsetT) *RoomUpdates {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &RoomUpdates{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *RoomUpdates) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *RoomUpdates) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *RoomUpdates) Updates(obj *RoomUpdate, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *RoomUpdates) UpdatesLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func RoomUpdatesStart(builder *flatbuffers.Builder) {
	builder.StartObject(1)
}
func RoomUpdatesAddUpdates(builder *flatbuffers.Builder, updates flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(updates), 0)
}
func RoomUpdatesStartUpdatesVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func RoomUpdatesEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}