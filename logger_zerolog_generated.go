// Code generated by ./cmd/generate-zerolog; DO NOT EDIT.

package diag

import (
	"fmt"
	"net"
	"time"
)

// MsgData fields functions

func (d *zerologLogData) Str(key string, value string) MsgData {
	return &zerologLogData{Event: d.Event.Str(key, value)}
}

func (d *zerologLogData) Strs(key string, value []string) MsgData {
	return &zerologLogData{Event: d.Event.Strs(key, value)}
}

func (d *zerologLogData) Stringer(key string, value fmt.Stringer) MsgData {
	return &zerologLogData{Event: d.Event.Stringer(key, value)}
}

func (d *zerologLogData) Bytes(key string, value []byte) MsgData {
	return &zerologLogData{Event: d.Event.Bytes(key, value)}
}

func (d *zerologLogData) Hex(key string, value []byte) MsgData {
	return &zerologLogData{Event: d.Event.Hex(key, value)}
}

func (d *zerologLogData) RawJSON(key string, value []byte) MsgData {
	return &zerologLogData{Event: d.Event.RawJSON(key, value)}
}

func (d *zerologLogData) Bool(key string, value bool) MsgData {
	return &zerologLogData{Event: d.Event.Bool(key, value)}
}

func (d *zerologLogData) Bools(key string, value []bool) MsgData {
	return &zerologLogData{Event: d.Event.Bools(key, value)}
}

func (d *zerologLogData) Int(key string, value int) MsgData {
	return &zerologLogData{Event: d.Event.Int(key, value)}
}

func (d *zerologLogData) Ints(key string, value []int) MsgData {
	return &zerologLogData{Event: d.Event.Ints(key, value)}
}

func (d *zerologLogData) Int8(key string, value int8) MsgData {
	return &zerologLogData{Event: d.Event.Int8(key, value)}
}

func (d *zerologLogData) Ints8(key string, value []int8) MsgData {
	return &zerologLogData{Event: d.Event.Ints8(key, value)}
}

func (d *zerologLogData) Int16(key string, value int16) MsgData {
	return &zerologLogData{Event: d.Event.Int16(key, value)}
}

func (d *zerologLogData) Ints16(key string, value []int16) MsgData {
	return &zerologLogData{Event: d.Event.Ints16(key, value)}
}

func (d *zerologLogData) Int32(key string, value int32) MsgData {
	return &zerologLogData{Event: d.Event.Int32(key, value)}
}

func (d *zerologLogData) Ints32(key string, value []int32) MsgData {
	return &zerologLogData{Event: d.Event.Ints32(key, value)}
}

func (d *zerologLogData) Int64(key string, value int64) MsgData {
	return &zerologLogData{Event: d.Event.Int64(key, value)}
}

func (d *zerologLogData) Ints64(key string, value []int64) MsgData {
	return &zerologLogData{Event: d.Event.Ints64(key, value)}
}

func (d *zerologLogData) Uint(key string, value uint) MsgData {
	return &zerologLogData{Event: d.Event.Uint(key, value)}
}

func (d *zerologLogData) Uints(key string, value []uint) MsgData {
	return &zerologLogData{Event: d.Event.Uints(key, value)}
}

func (d *zerologLogData) Uint8(key string, value uint8) MsgData {
	return &zerologLogData{Event: d.Event.Uint8(key, value)}
}

func (d *zerologLogData) Uints8(key string, value []uint8) MsgData {
	return &zerologLogData{Event: d.Event.Uints8(key, value)}
}

func (d *zerologLogData) Uint16(key string, value uint16) MsgData {
	return &zerologLogData{Event: d.Event.Uint16(key, value)}
}

func (d *zerologLogData) Uints16(key string, value []uint16) MsgData {
	return &zerologLogData{Event: d.Event.Uints16(key, value)}
}

func (d *zerologLogData) Uint32(key string, value uint32) MsgData {
	return &zerologLogData{Event: d.Event.Uint32(key, value)}
}

func (d *zerologLogData) Uints32(key string, value []uint32) MsgData {
	return &zerologLogData{Event: d.Event.Uints32(key, value)}
}

func (d *zerologLogData) Uint64(key string, value uint64) MsgData {
	return &zerologLogData{Event: d.Event.Uint64(key, value)}
}

func (d *zerologLogData) Uints64(key string, value []uint64) MsgData {
	return &zerologLogData{Event: d.Event.Uints64(key, value)}
}

func (d *zerologLogData) Float32(key string, value float32) MsgData {
	return &zerologLogData{Event: d.Event.Float32(key, value)}
}

func (d *zerologLogData) Floats32(key string, value []float32) MsgData {
	return &zerologLogData{Event: d.Event.Floats32(key, value)}
}

func (d *zerologLogData) Float64(key string, value float64) MsgData {
	return &zerologLogData{Event: d.Event.Float64(key, value)}
}

func (d *zerologLogData) Floats64(key string, value []float64) MsgData {
	return &zerologLogData{Event: d.Event.Floats64(key, value)}
}

func (d *zerologLogData) Time(key string, value time.Time) MsgData {
	return &zerologLogData{Event: d.Event.Time(key, value)}
}

func (d *zerologLogData) Times(key string, value []time.Time) MsgData {
	return &zerologLogData{Event: d.Event.Times(key, value)}
}

func (d *zerologLogData) IPAddr(key string, value net.IP) MsgData {
	return &zerologLogData{Event: d.Event.IPAddr(key, value)}
}

func (d *zerologLogData) IPPrefix(key string, value net.IPNet) MsgData {
	return &zerologLogData{Event: d.Event.IPPrefix(key, value)}
}

func (d *zerologLogData) MACAddr(key string, value net.HardwareAddr) MsgData {
	return &zerologLogData{Event: d.Event.MACAddr(key, value)}
}

func (d *zerologLogData) Interface(key string, value interface{}) MsgData {
	return &zerologLogData{Event: d.Event.Interface(key, value)}
}
