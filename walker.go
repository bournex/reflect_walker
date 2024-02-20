package reflect_walker

import (
	"context"
	"reflect"
)

type Node_routine func(ctx context.Context, node TreeNode)
type Walker interface {
	Walk(context.Context, interface{}) interface{}
}

const NoDepthLimit = -1
const DepthCtxKey = "_reflect_walker_curr_depth"

type WalkOption func(tw *walker)

func WithMaxDepth(max_depth int) WalkOption {
	return func(tw *walker) {
		tw.maxDepth = max_depth
	}
}

// 顺序相关
func WithRoutine(routines ...Node_routine) WalkOption {
	return func(tw *walker) {
		tw.routines = routines
	}
}

// 当结构体、interface等类型作为map的key时，map无法做json序列化，json.Marshal会报"json: unsupported type..."的错误
// 若设置jsonable为true，就会尝试在遍历过程中将map的key转换为字符串类型，使遍历过后的结果可以直接被json序列化
func WithJsonableMap() WalkOption {
	return func(tw *walker) {
		tw.jsonable = true
	}
}

func NewTreeWalker(wo ...WalkOption) Walker {
	tw := &walker{maxDepth: NoDepthLimit}
	for _, option := range wo {
		option(tw)
	}
	return tw
}

type walker struct {
	maxDepth      int            // recursive depth
	jsonable      bool           // make input json marshalable or let it be
	forceOverride bool           // copy or in-place override
	routines      []Node_routine // custom callback routine
}

func (tr *walker) Walk(ctx context.Context, in interface{}) interface{} {
	if in == nil {
		return in
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return tr.walk(ctx, in)
}

func (tr *walker) walk(ctx context.Context, in interface{}) interface{} {
	if nctx, too_deep := tr.dive(ctx); too_deep {
		return in
	} else {
		ctx = nctx
	}
	defer tr.rise(ctx)

	intyp := reflect.TypeOf(in)
	// inval := reflect.ValueOf(in)

	switch intyp.Kind() {
	case reflect.Map:
		in = tr.walk_map(ctx, in)
	case reflect.Slice:
		in = tr.walk_slice(ctx, in)
	case reflect.Struct:
		in = tr.walk_struct(ctx, in)
	case reflect.Pointer:
		in = tr.walk_pointer(ctx, in)
	case reflect.Interface:
		// do nothing
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, // 有符号数
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, // 无符号数
		reflect.Float32, reflect.Float64, // 浮点数
		reflect.Complex64, reflect.Complex128, // 复数
		reflect.String, // 字符串
		reflect.Bool:   // 布尔
		in = tr.walk_literal(ctx, in, false)
	default:
	}
	return in
}

func (tr *walker) walk_literal(ctx context.Context, in interface{}, settable bool) interface{} {
	intyp := reflect.TypeOf(in)
	inval := reflect.ValueOf(in)
	if settable {
		intyp = intyp.Elem()
		inval = inval.Elem()
	}

	node := &treeNode{
		nType: NodeType_literal,
	}
	node.nValue = &treeVariable{node: node, t: intyp, value: inval.Interface()}
	var override bool

	for _, r := range tr.routines {
		r(ctx, node)
		rt := node.getAction()

		if rt == routine_override {
			override = true
		}
	}

	if override {
		if settable {
			inval.Set(node.nValue.rvalue())
		} else {
			nval := reflect.New(intyp).Elem()
			newVal := reflect.ValueOf(node.nValue.Interface())
			nval.Set(newVal)
			in = nval.Interface()
		}
	}

	return in
}

func (tr *walker) walk_slice(ctx context.Context, in interface{}) interface{} {
	intyp := reflect.TypeOf(in)
	inval := reflect.ValueOf(in)

	// 排除掉有类型信息的nil值
	if inval.IsNil() {
		return in
	}

	mdval := reflect.MakeSlice(intyp, 0, inval.Cap())

	for i := 0; i < inval.Len(); i++ {

		val := inval.Index(i)
		val = tr.unpack_value(val)

		if !tr.is_literal(&val) {
			val = reflect.ValueOf(tr.walk(ctx, val.Interface()))
			mdval = reflect.Append(mdval, val)
			continue
		}

		node := &treeNode{
			nType: NodeType_slice_member,
		}
		node.nValue = &treeVariable{node: node, t: val.Type(), value: val.Interface()}

		override := false
		var rt routine_action
		for _, r := range tr.routines {
			r(ctx, node)

			rt = node.getAction()
			if rt == routine_delete {
				break
			} else if rt == routine_override {
				override = true
			}
		}

		if rt == routine_delete {
			continue
		}

		if override {
			val = reflect.ValueOf(node.nValue.Interface())
		}

		mdval = reflect.Append(mdval, val)
	}
	return mdval.Interface()
}

func (tr *walker) walk_map(ctx context.Context, in interface{}) interface{} {
	intyp := reflect.TypeOf(in)
	inval := reflect.ValueOf(in)

	// 排除掉有类型信息的nil值
	if inval.IsNil() {
		return in
	}

	var (
		walkmap     reflect.Value
		walkmapType reflect.Type = intyp
	)

	if tr.jsonable {
		walkmapType = reflect.MapOf(reflect.TypeOf(""), intyp.Elem())
	}
	walkmap = reflect.MakeMapWithSize(walkmapType, inval.Len())

	iter := inval.MapRange()
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		key = tr.unpack_value(key)
		val = tr.unpack_value(val)

		if !tr.is_literal(&val) {
			val = reflect.ValueOf(tr.walk(ctx, val.Interface()))
		}

		node := &treeNode{
			nType: NodeType_map_pair,
		}
		node.nKey = &treeVariable{node: node, t: key.Type(), value: key.Interface()}
		node.nValue = &treeVariable{node: node, t: val.Type(), value: val.Interface()}

		var (
			rt       routine_action
			override bool
		)

		for _, r := range tr.routines {
			r(ctx, node)

			rt = node.getAction()
			if rt == routine_delete {
				break
			} else if rt == routine_override {
				override = true
			}
		}

		if rt == routine_delete {
			continue
		}

		if override {
			key = reflect.ValueOf(node.nKey.Interface())
			val = reflect.ValueOf(node.nValue.Interface())
		}

		walkmap.SetMapIndex(key, val)
	}
	return walkmap.Interface()
}

func (tr *walker) walk_struct(ctx context.Context, in interface{}) interface{} {
	intyp := reflect.TypeOf(in)
	inval := reflect.ValueOf(in)

	kind := intyp.Kind()
	writable := kind == reflect.Pointer || kind == reflect.Interface
	if writable {
		if inval.IsNil() {
			return in
		}
		inval = inval.Elem()
		intyp = intyp.Elem()
	}

	for i := 0; i < inval.NumField(); i++ {
		val := inval.Field(i)
		typ := intyp.Field(i)

		if !val.CanInterface() {
			// 只walk公有成员
			continue
		}

		if !tr.is_literal(&val) {
			val = reflect.ValueOf(tr.walk(ctx, val.Interface()))

			if writable {
				inval.Field(i).Set(val)
			}

			continue
		}

		node := &treeNode{
			nType: NodeType_struct_member,
		}
		node.nKey = &treeVariable{node: node, t: reflect.TypeOf(""), value: typ.Name}
		node.nValue = &treeVariable{node: node, t: val.Type(), value: val.Interface()}

		override := false
		var rt routine_action
		for _, r := range tr.routines {
			r(ctx, node)

			rt = node.getAction()
			if rt == routine_delete {
				// struct成员不支持delete，与blank效果一样
				break
			} else if rt == routine_override {
				override = true
			}
		}

		if override && writable {
			val = reflect.ValueOf(node.nValue.Interface())
			inval.Field(i).Set(val)
		}
	}
	// -
	return in
}

func (tr *walker) walk_pointer(ctx context.Context, in interface{}) interface{} {
	intyp := reflect.TypeOf(in)
	inval := reflect.ValueOf(in)

	typ := intyp.Elem()
	if typ.Kind() == reflect.Map {
		in = tr.walk_map(ctx, in)
	} else if typ.Kind() == reflect.Slice {
		in = tr.walk_slice(ctx, in)
	} else if typ.Kind() == reflect.Struct {
		in = tr.walk_struct(ctx, in)
	} else {
		switch typ.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, // 有符号数
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, // 无符号数
			reflect.Float32, reflect.Float64, // 浮点数
			reflect.Complex64, reflect.Complex128, // 复数
			reflect.String, // 字符串
			reflect.Bool:   // 布尔
			in = tr.walk_literal(ctx, inval.Interface(), true)
		}
		// in = tr.walk(ctx, inval.Elem().Interface())
	}

	return in
}

func (tr *walker) is_literal(val *reflect.Value) bool {
	if val == nil {
		return true
	}
	switch val.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Struct, reflect.Ptr:
		return false
	}
	return true
}

func (tr *walker) unpack_value(val reflect.Value) reflect.Value {
	if (val.Type().Kind() == reflect.Pointer || val.Type().Kind() == reflect.Interface) && !val.IsNil() {
		return val.Elem()
	}
	return val
}

func (tr *walker) dive(ctx context.Context) (context.Context, bool) {
	if tr.maxDepth == NoDepthLimit {
		return ctx, false
	}

	depth := ctx.Value(DepthCtxKey)
	if depth == nil {
		return context.WithValue(ctx, DepthCtxKey, new(int)), false
	}

	if n, ok := depth.(*int); ok {
		curDepth := *n
		if curDepth < tr.maxDepth {
			*n = curDepth + 1
		} else {
			return ctx, true
		}
	}

	return ctx, false
}

func (tr *walker) rise(ctx context.Context) {
	if tr.maxDepth == NoDepthLimit {
		return
	}

	n, _ := ctx.Value(DepthCtxKey).(*int)
	curDepth := *n
	if curDepth > 0 {
		*n = curDepth - 1
	}
}
