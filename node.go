package reflect_walker

import (
	"errors"
	"reflect"
)

var (
	ErrTypeAssertFailed = errors.New("type assertion failed")
)

// TreeNode类型
type nType int

const (
	// nType枚举
	NodeType_map_pair      = iota // map key
	NodeType_slice_member         // slice items
	NodeType_struct_member        // struct
	NodeType_literal              // literal
)

type routine_action int

const (
	// Routine_actions
	routine_blank    = iota // 什么也不做，用于读遍历
	routine_override        // 覆盖回写，用于自定义修改
	routine_delete          // 删除字段，仅对node类型为NodeType_key的有效
)

type TreeVariable interface {
	// NType
	typetype() reflect.Type
	TypeName() string       // 原始类型名
	TypeKind() reflect.Kind // 原始Kind

	Interface() interface{}

	// 类型断言
	String() (string, error)
	Int() (int, error)
	Int8() (int8, error)
	Int16() (int16, error)
	Int32() (int32, error)
	Int64() (int64, error)
	Uint() (uint, error)
	Uint8() (uint8, error)
	Uint16() (uint16, error)
	Uint32() (uint32, error)
	Uint64() (uint64, error)
	Float32() (float32, error)
	Float64() (float64, error)
	Bool() (bool, error)

	// 类型断言，类型异常会panic
	MustString() string
	MustInt() int
	MustInt8() int8
	MustInt16() int16
	MustInt32() int32
	MustInt64() int64
	MustUint() uint
	MustUint8() uint8
	MustUint16() uint16
	MustUint32() uint32
	MustUint64() uint64
	MustFloat32() float32
	MustFloat64() float64
	MustBool() bool

	// 修改值
	Set(variable interface{})
}

type treeVariable struct {
	node  TreeNode
	t     reflect.Type
	value interface{}
}

// 获取反射类型
func (tv *treeVariable) typetype() reflect.Type {
	return tv.t
}

// 获取原始类型名
func (tv *treeVariable) TypeName() string {
	return tv.t.Name()
}

// 获取反射Kind
func (tv *treeVariable) TypeKind() reflect.Kind {
	return tv.t.Kind()
}

func (tv *treeVariable) Interface() interface{} {
	return tv.value
}

func (tv *treeVariable) String() (string, error) {
	if s, e := tv.value.(string); e {
		return s, nil
	}
	return "", ErrTypeAssertFailed
}

func (tv *treeVariable) Int() (int, error) {
	if u8, e := tv.value.(int); e {
		return u8, nil
	}
	return 0, ErrTypeAssertFailed
}

func (tv *treeVariable) Int8() (int8, error) {
	if i8, e := tv.value.(int8); e {
		return i8, nil
	}
	return 0, ErrTypeAssertFailed
}

func (tv *treeVariable) Int16() (int16, error) {
	if i16, e := tv.value.(int16); e {
		return i16, nil
	}
	return 0, ErrTypeAssertFailed
}

func (tv *treeVariable) Int32() (int32, error) {
	if i32, e := tv.value.(int32); e {
		return i32, nil
	}
	return 0, ErrTypeAssertFailed
}

func (tv *treeVariable) Int64() (int64, error) {
	if i64, e := tv.value.(int64); e {
		return i64, nil
	}
	return 0, ErrTypeAssertFailed
}

func (tv *treeVariable) Uint() (uint, error) {
	if u8, e := tv.value.(uint); e {
		return u8, nil
	}
	return 0, ErrTypeAssertFailed
}

func (tv *treeVariable) Uint8() (uint8, error) {
	if u8, e := tv.value.(uint8); e {
		return u8, nil
	}
	return 0, ErrTypeAssertFailed
}

func (tv *treeVariable) Uint16() (uint16, error) {
	if u16, e := tv.value.(uint16); e {
		return u16, nil
	}
	return 0, ErrTypeAssertFailed
}

func (tv *treeVariable) Uint32() (uint32, error) {
	if u32, e := tv.value.(uint32); e {
		return u32, nil
	}
	return 0, ErrTypeAssertFailed
}

func (tv *treeVariable) Uint64() (uint64, error) {
	if u64, e := tv.value.(uint64); e {
		return u64, nil
	}
	return 0, ErrTypeAssertFailed
}

func (tv *treeVariable) Float32() (float32, error) {
	if f32, e := tv.value.(float32); e {
		return f32, nil
	}
	return 0, ErrTypeAssertFailed
}

func (tv *treeVariable) Float64() (float64, error) {
	if f64, e := tv.value.(float64); e {
		return f64, nil
	}
	return 0, ErrTypeAssertFailed
}

func (tv *treeVariable) Bool() (bool, error) {
	if b, e := tv.value.(bool); e {
		return b, nil
	}
	return false, ErrTypeAssertFailed
}

func (tv *treeVariable) MustString() string {
	return tv.value.(string)
}

func (tv *treeVariable) MustInt() int {
	return tv.value.(int)
}

func (tv *treeVariable) MustInt8() int8 {
	return tv.value.(int8)
}

func (tv *treeVariable) MustInt16() int16 {
	return tv.value.(int16)
}

func (tv *treeVariable) MustInt32() int32 {
	return tv.value.(int32)
}

func (tv *treeVariable) MustInt64() int64 {
	return tv.value.(int64)
}

func (tv *treeVariable) MustUint() uint {
	return tv.value.(uint)
}

func (tv *treeVariable) MustUint8() uint8 {
	return tv.value.(uint8)
}

func (tv *treeVariable) MustUint16() uint16 {
	return tv.value.(uint16)
}

func (tv *treeVariable) MustUint32() uint32 {
	return tv.value.(uint32)
}

func (tv *treeVariable) MustUint64() uint64 {
	return tv.value.(uint64)
}

func (tv *treeVariable) MustFloat32() float32 {
	return tv.value.(float32)
}

func (tv *treeVariable) MustFloat64() float64 {
	return tv.value.(float64)
}

func (tv *treeVariable) MustBool() bool {
	return tv.value.(bool)
}

func (tv *treeVariable) Set(value interface{}) {
	n := tv.node
	if n.getAction() == routine_delete {
		return
	}
	tv.node.setAction(routine_override)
	tv.value = value
}

type TreeNode interface {
	Type() nType
	Key() TreeVariable
	Value() TreeVariable
	Delete()
	// 内部接口
	getAction() routine_action
	setAction(routine_action)
}

type treeNode struct {
	nType               // 节点类型
	nKey   TreeVariable // 节点索引，当前仅map类型
	nValue TreeVariable // 节点值
	action routine_action
}

func (tn *treeNode) Type() nType {
	return tn.nType
}

func (tn *treeNode) Key() TreeVariable {
	return tn.nKey
}

func (tn *treeNode) Value() TreeVariable {
	return tn.nValue
}

func (tn *treeNode) Delete() {
	tn.action = routine_delete
}

func (tn *treeNode) getAction() routine_action {
	return tn.action
}

func (tn *treeNode) setAction(action routine_action) {
	if tn.action == routine_delete {
		return
	}
	tn.action = action
}
