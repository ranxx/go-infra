package message

import "reflect"

// V 消息实例信息接口
type V interface {
	GetV() interface{}
	GetT() reflect.Type
	GetName() string
}

// Msger 消息编解码接口
type Msger interface {
	// 注册消息（使用默认分组）
	Register(msg interface{}) uint16

	// 创建新的消息分组，返回 Msger 以支持链式调用
	NewGroup(name string, begin uint16) Msger

	// 在指定分组中注册消息
	RegisterByGroup(group string, msg interface{}) uint16

	// 解码：字节流 -> 消息对象
	Unmarshal(data []byte) (interface{}, error)

	// 编码：消息对象 -> 字节流
	Marshal(msg interface{}) ([]byte, error)

	// 遍历所有已注册的消息
	Range(f func(id uint16, v V))

	// 通过协议号获取消息实例
	GetMessageByID(id uint16) interface{}

	// 通过消息实例获取协议号
	GetIDByMessage(v interface{}) int
}

// msgInfo 消息注册信息
type msgInfo struct {
	t reflect.Type
	v interface{}
	n string
}

func (m *msgInfo) GetV() interface{}  { return m.v }
func (m *msgInfo) GetT() reflect.Type { return m.t }
func (m *msgInfo) GetName() string    { return m.n }

// groupInfo 分组信息
type groupInfo struct {
	name    string
	begin   uint16
	current uint16
}

// DefaultGroupName 默认分组名称
const DefaultGroupName = "default"

var _default = NewProtobuf(false)
var _defaultJSON = NewJSON(false)

// Default 返回默认 Protobuf 消息处理器
func Default() Msger { return _default }

// DefaultJSON 返回默认 JSON 消息处理器
func DefaultJSON() Msger { return _defaultJSON }
