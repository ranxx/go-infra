package message

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
)

type jsonMessage struct {
	littleEndian bool
	msgInfo      map[uint16]*msgInfo // 改为 map
	msgIds       map[reflect.Type]uint16
	groups       map[string]*groupInfo // 分组信息
}

// NewJSON creates a new JSON-based message handler
func NewJSON(littleEndian bool) Msger {
	j := &jsonMessage{
		littleEndian: littleEndian,
		msgInfo:      make(map[uint16]*msgInfo),
		msgIds:       map[reflect.Type]uint16{},
		groups:       make(map[string]*groupInfo),
	}

	// 创建默认分组
	j.groups[DefaultGroupName] = &groupInfo{
		name:    DefaultGroupName,
		begin:   0,
		current: 0,
	}

	return j
}

// Register 注册消息（使用默认分组）
func (j *jsonMessage) Register(msg interface{}) uint16 {
	return j.RegisterByGroup(DefaultGroupName, msg)
}

// Unmarshal 解码消息
func (j *jsonMessage) Unmarshal(data []byte) (interface{}, error) {
	if len(data) < 2 {
		return nil, errors.New("json data too short")
	}

	// 读取消息ID
	var id uint16
	if j.littleEndian {
		id = binary.LittleEndian.Uint16(data)
	} else {
		id = binary.BigEndian.Uint16(data)
	}

	info, exists := j.msgInfo[id]
	if !exists {
		return nil, fmt.Errorf("message id %v not registered", id)
	}

	// 创建消息实例
	msg := reflect.New(info.t.Elem()).Interface()

	// JSON反序列化
	return msg, json.Unmarshal(data[2:], msg)
}

// Marshal 编码消息
func (j *jsonMessage) Marshal(msg interface{}) ([]byte, error) {
	t := reflect.TypeOf(msg)

	// 获取消息ID
	_id, ok := j.msgIds[t]
	if !ok {
		return nil, fmt.Errorf("message %s not registered", t)
	}

	// JSON序列化
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// 构造最终数据：ID(2字节) + JSON数据
	result := make([]byte, 2, 2+len(data))
	if j.littleEndian {
		binary.LittleEndian.PutUint16(result, _id)
	} else {
		binary.BigEndian.PutUint16(result, _id)
	}
	result = append(result, data...)

	return result, nil
}

// Range 遍历所有注册的消息
func (j *jsonMessage) Range(f func(id uint16, v V)) {
	// 收集所有ID并排序
	var ids []uint16
	for id := range j.msgInfo {
		ids = append(ids, id)
	}

	// 使用标准库排序（性能更好）
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	// 按顺序调用回调函数
	for _, id := range ids {
		if info, exists := j.msgInfo[id]; exists {
			f(id, info)
		}
	}
}

// GetMessageByID 通过协议号获取消息
func (j *jsonMessage) GetMessageByID(id uint16) interface{} {
	info, exists := j.msgInfo[id]
	if !exists {
		return nil
	}
	return info.v
}

// GetIDByMessage 通过消息获取协议号
func (j *jsonMessage) GetIDByMessage(v interface{}) int {
	t := reflect.TypeOf(v)
	_id, ok := j.msgIds[t]
	if !ok {
		return -1
	}
	return int(_id)
}

// NewGroup 创建新的消息分组
func (j *jsonMessage) NewGroup(name string, begin uint16) Msger {
	if _, exists := j.groups[name]; exists {
		panic(fmt.Sprintf("group %s already exists", name))
	}

	j.groups[name] = &groupInfo{
		name:    name,
		begin:   begin,
		current: begin,
	}

	return j
}

// RegisterByGroup 在指定分组中注册消息
func (j *jsonMessage) RegisterByGroup(group string, msg interface{}) uint16 {
	groupInfo, exists := j.groups[group]
	if !exists {
		panic(fmt.Sprintf("group %s not found", group))
	}

	t := reflect.TypeOf(msg)
	if t == nil || t.Kind() != reflect.Ptr {
		panic("msg must be a pointer")
	}

	if _, ok := j.msgIds[t]; ok {
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

	// 在指定位置设置消息信息
	j.msgInfo[groupInfo.current] = &info
	j.msgIds[t] = groupInfo.current

	// 返回当前ID并递增
	id := groupInfo.current
	groupInfo.current++

	return id
}
