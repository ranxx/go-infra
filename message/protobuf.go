package message

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"
)

type protobuf struct {
	littleEndian bool
	msgInfo      map[uint16]*msgInfo
	msgIds       map[reflect.Type]uint16
	groups       map[string]*groupInfo
}

// NewProtobuf 创建 Protobuf 消息处理器
func NewProtobuf(littleEndian bool) Msger {
	p := &protobuf{
		littleEndian: littleEndian,
		msgInfo:      make(map[uint16]*msgInfo),
		msgIds:       map[reflect.Type]uint16{},
		groups:       make(map[string]*groupInfo),
	}

	p.groups[DefaultGroupName] = &groupInfo{
		name:    DefaultGroupName,
		begin:   0,
		current: 0,
	}

	return p
}

func (p *protobuf) SetLittleEndian(ok bool) {
	p.littleEndian = ok
}

func (p *protobuf) Register(msg interface{}) uint16 {
	return p.RegisterByGroup(DefaultGroupName, msg)
}

func (p *protobuf) Unmarshal(data []byte) (interface{}, error) {
	if len(data) < 2 {
		return nil, errors.New("protobuf data too short")
	}

	var id uint16
	if p.littleEndian {
		id = binary.LittleEndian.Uint16(data)
	} else {
		id = binary.BigEndian.Uint16(data)
	}

	info, exists := p.msgInfo[id]
	if !exists {
		return nil, fmt.Errorf("message id %v not registered", id)
	}

	msg := reflect.New(info.t.Elem()).Interface()
	return msg, proto.Unmarshal(data[2:], msg.(proto.Message))
}

func (p *protobuf) Marshal(msg interface{}) ([]byte, error) {
	t := reflect.TypeOf(msg)

	_id, ok := p.msgIds[t]
	if !ok {
		return nil, fmt.Errorf("message %s not registered", t)
	}

	data, err := proto.Marshal(msg.(proto.Message))

	id := make([]byte, 2, 2+len(data))
	if p.littleEndian {
		binary.LittleEndian.PutUint16(id, _id)
	} else {
		binary.BigEndian.PutUint16(id, _id)
	}
	id = append(id, data...)
	return id, err
}

func (p *protobuf) Range(f func(id uint16, v V)) {
	var ids []uint16
	for id := range p.msgInfo {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	for _, id := range ids {
		if info, exists := p.msgInfo[id]; exists {
			f(id, info)
		}
	}
}

func (p *protobuf) GetMessageByID(id uint16) interface{} {
	info, exists := p.msgInfo[id]
	if !exists {
		return nil
	}
	return info.v
}

func (p *protobuf) GetIDByMessage(v interface{}) int {
	t := reflect.TypeOf(v)
	_id, ok := p.msgIds[t]
	if !ok {
		return -1
	}
	return int(_id)
}

func (p *protobuf) NewGroup(name string, begin uint16) Msger {
	if _, exists := p.groups[name]; exists {
		panic(fmt.Sprintf("group %s already exists", name))
	}

	p.groups[name] = &groupInfo{
		name:    name,
		begin:   begin,
		current: begin,
	}

	return p
}

func (p *protobuf) RegisterByGroup(group string, msg interface{}) uint16 {
	groupInfo, exists := p.groups[group]
	if !exists {
		panic(fmt.Sprintf("group %s not found", group))
	}

	t := reflect.TypeOf(msg)
	if t == nil || t.Kind() != reflect.Ptr {
		panic("msg must be a pointer")
	}
	_, ok := msg.(proto.Message)
	if !ok {
		panic("msg must be protobuf")
	}

	if _, ok := p.msgIds[t]; ok {
		panic(fmt.Sprintf("msg %s is already registered", t))
	}

	if groupInfo.current == math.MaxUint16 {
		panic(fmt.Sprintf("group %s has too many messages (max = %v)", group, math.MaxUint16))
	}

	names := strings.Split(t.String(), ".")

	info := msgInfo{
		t: t,
		v: msg,
		n: names[len(names)-1],
	}

	p.msgInfo[groupInfo.current] = &info
	p.msgIds[t] = groupInfo.current

	id := groupInfo.current
	groupInfo.current++

	return id
}
