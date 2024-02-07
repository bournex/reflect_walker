package reflect_walker

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

func Test_Literal(t *testing.T) {
	testCases := []struct {
		name   string
		input  interface{}
		expect interface{}
	}{
		{
			name:   "字面量测试-nil",
			input:  nil,
			expect: nil,
		},
		{
			name:   "字面量测试-字符串",
			input:  "hello",
			expect: "hello",
		},
		{
			name:   "字面量测试-整数",
			input:  5,
			expect: 5,
		},
		{
			name:   "字面量测试-浮点数",
			input:  3.14,
			expect: 3.14,
		},
		{
			name: "聚合基础类型测试",
			input: map[interface{}]interface{}{
				"foo":   "bar",         // string
				"n1":    1,             // int
				"n2":    int8(1),       // int8
				"n3":    int16(1),      // int16
				"n4":    int32(1),      // int32
				"n5":    int64(1),      // int64
				"n6":    uint(1),       // uint
				"n7":    uint8(1),      // uint8
				"n8":    uint16(1),     // uint16
				"n9":    uint32(1),     // uint32
				"n10":   uint64(1),     // uint64
				"alice": float32(3.14), // float32
				"bob":   float64(3.14), // float64
			},
			expect: map[interface{}]interface{}{
				"foo":   "bar",         // string
				"n1":    1,             // int
				"n2":    int8(1),       // int8
				"n3":    int16(1),      // int16
				"n4":    int32(1),      // int32
				"n5":    int64(1),      // int64
				"n6":    uint(1),       // uint
				"n7":    uint8(1),      // uint8
				"n8":    uint16(1),     // uint16
				"n9":    uint32(1),     // uint32
				"n10":   uint64(1),     // uint64
				"alice": float32(3.14), // float32
				"bob":   float64(3.14), // float64
			},
		},
	}
	for _, v := range testCases {
		t.Run(v.name, func(t *testing.T) {
			tw := NewTreeWalker()

			got := tw.Walk(context.Background(), v.input)
			if !reflect.DeepEqual(got, v.expect) {
				t.Errorf("miss match: \n\tinput:  %+v\n\texpect:%+v\n\tgot:   %+v", v.input, v.expect, got)
			}
		})
	}
}

// 由于float类型不能用reflect.DeepEqual比较，为float类型单独实现单测
func Test_Float(t *testing.T) {
	testCases := []struct {
		name    string
		input   []interface{}
		routine []Node_routine
		expect  []interface{}
	}{
		{
			name: "对比float",
			input: []interface{}{
				float32(3.14),
				float64(3.14),
			},
			routine: []Node_routine{
				func(ctx context.Context, node TreeNode) {
					if node.Type() == NodeType_slice_member {
						switch node.Value().TypeKind() {
						case reflect.Float32:
							f, _ := node.Value().Float32()
							node.Value().Set(f * 10)
						case reflect.Float64:
							f, _ := node.Value().Float64()
							node.Value().Set(f * 10)
						}
					}
				},
			},
			expect: []interface{}{
				float32(31.4),
				float64(31.4),
			},
		},
	}
	for _, v := range testCases {
		t.Run(v.name, func(t *testing.T) {
			tw := NewTreeWalker(
				WithRoutine(v.routine...),
			)

			got := tw.Walk(context.Background(), v.input)
			g, _ := got.([]interface{})
			if len(g) != len(v.expect) {
				t.Errorf("miss match: \n\tinput:  %+v\n\texpect:%+v\n\tgot:   %+v", v.input, v.expect, got)
				return
			}

			g0 := g[0]
			e0 := v.expect[0]
			gf0 := g0.(float32)
			ef0 := e0.(float32)

			float32epsilon := 0.00001
			if math.Abs(float64(gf0-ef0)) >= float32epsilon {
				t.Errorf("float32 miss match: \n\tinput:  %+v\n\texpect:%+v\n\tgot:   %+v", v.input, v.expect, got)
			}

			g1 := g[1]
			e1 := v.expect[1]
			gf1 := g1.(float64)
			ef1 := e1.(float64)

			float64epsilon := 0.000000001
			if math.Abs(gf1-ef1) >= float64epsilon {
				t.Errorf("float64 miss match: \n\tinput:  %+v\n\texpect:%+v\n\tgot:   %+v", v.input, v.expect, got)
			}
		})
	}
}

func Test_Map(t *testing.T) {

	testCases := []struct {
		name     string
		input    interface{}
		routines []Node_routine
		expect   interface{}
	}{
		{
			name: "已知类型map测试",
			input: map[string]int{
				"alice": 5,
				"bob":   3,
			},
			expect: map[string]int{
				"alice": 5,
				"bob":   3,
			},
		},
		{
			name: "嵌套map测试",
			input: map[interface{}]interface{}{
				"alice": map[interface{}]interface{}{
					"age": 5,
				},
				"bob": map[interface{}]interface{}{
					"age": 3.14,
				},
			},
			expect: map[string]interface{}{
				"alice": map[string]interface{}{
					"age": 5,
				},
				"bob": map[string]interface{}{
					"age": 3.14,
				},
			},
		},
		{
			name:   "空map测试1",
			input:  map[interface{}]interface{}{},
			expect: map[string]interface{}{},
		},
		{
			name:   "空map测试2",
			input:  map[string]interface{}{},
			expect: map[string]interface{}{},
		},
		{
			name:   "空map测试3",
			input:  map[interface{}]int{},
			expect: map[string]int{},
		},
		{
			name:   "空map测试4",
			input:  map[string]int{},
			expect: map[string]int{},
		},

		{
			name: "字面量测试",
			input: map[interface{}]interface{}{
				"alice": 5,
				"bob":   3.14,
			},
			expect: map[string]interface{}{
				"alice": 5,
				"bob":   3.14,
			},
		},
	}
	for _, v := range testCases {
		t.Run(v.name, func(t *testing.T) {
			tw := NewTreeWalker(
				WithRoutine(v.routines...),
				WithJsonableMap(),
			)
			got := tw.Walk(context.Background(), v.input)
			if !reflect.DeepEqual(got, v.expect) {
				t.Errorf("miss match: \n\tinput:  %+v\n\texpect:%+v\n\tgot:   %+v", v.input, v.expect, got)
			}
		})
	}
}

func Test_Slice(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		routines []Node_routine
		expect   interface{}
	}{
		{
			name:   "空slice测试",
			input:  []interface{}{},
			expect: []interface{}{},
		},
		{
			name: "字面量成员测试",
			input: []interface{}{
				5,
				"alice",
			},
			expect: []interface{}{
				5,
				"alice",
			},
		},
		{
			name: "嵌套两级slice测试",
			input: []interface{}{
				[]interface{}{
					1, 2, 3, 4, 5,
				},
				[]interface{}{"alice", "bob"},
			},
			expect: []interface{}{
				[]interface{}{
					1, 2, 3, 4, 5,
				},
				[]interface{}{"alice", "bob"},
			},
		},
	}
	for _, v := range testCases {
		t.Run(v.name, func(t *testing.T) {
			tw := NewTreeWalker(
				WithRoutine(v.routines...),
			)
			got := tw.Walk(context.Background(), v.input)
			if !reflect.DeepEqual(got, v.expect) {
				t.Errorf("miss match: \n\tinput:  %+v\n\texpect:%+v\n\tgot:   %+v", v.input, v.expect, got)
			}
		})
	}
}

func Test_map_routine(t *testing.T) {
	testCases := []struct {
		name     string
		enable   bool
		input    interface{}
		jsonable bool
		routines []Node_routine
		expect   interface{}
	}{
		{
			name: "替换覆盖map-key测试",
			input: map[interface{}]interface{}{
				"alice": map[interface{}]interface{}{
					"age": 5,
				},
				"bob": map[interface{}]interface{}{
					"firstname": "james",
				},
			},
			jsonable: true,
			routines: []Node_routine{
				func(ctx context.Context, node TreeNode) {
					if node.Type() == NodeType_map_pair {
						s, err := node.Key().String()
						if err != nil {
							fmt.Println(err)
							return
						}
						node.Key().Set(strings.ToUpper(s))
					}
				},
			},
			expect: map[string]interface{}{
				"ALICE": map[string]interface{}{
					"AGE": 5,
				},
				"BOB": map[string]interface{}{
					"FIRSTNAME": "james",
				},
			},
		},
		{
			name: "空操作map测试",
			input: map[interface{}]interface{}{
				"alice": map[interface{}]interface{}{
					"age": 5,
				},
				"bob": map[interface{}]interface{}{
					"firstname": "james",
				},
			},
			jsonable: true,
			routines: []Node_routine{
				func(ctx context.Context, node TreeNode) {
					return
				},
			},
			expect: map[string]interface{}{
				"alice": map[string]interface{}{
					"age": 5,
				},
				"bob": map[string]interface{}{
					"firstname": "james",
				},
			},
		},
		{
			name: "替换覆盖map-value(string)测试",
			input: map[interface{}]interface{}{
				"alice": map[interface{}]interface{}{
					"age": 5,
				},
				"bob": map[interface{}]interface{}{
					"firstname": "james",
				},
			},
			jsonable: true,
			routines: []Node_routine{
				func(ctx context.Context, node TreeNode) {
					if node.Type() == NodeType_map_pair {
						switch node.Value().TypeKind() {
						case reflect.String:
							s, _ := node.Value().String()
							node.Value().Set(strings.ToUpper(s))
						}
					}
				},
			},
			expect: map[string]interface{}{
				"alice": map[string]interface{}{
					"age": 5,
				},
				"bob": map[string]interface{}{
					"firstname": "JAMES",
				},
			},
		},
		{
			name: "替换覆盖map-value(int)测试",
			input: map[interface{}]interface{}{
				"alice": map[interface{}]interface{}{
					"age": 5,
				},
				"bob": map[interface{}]interface{}{
					"firstname": "james",
				},
			},
			jsonable: true,
			routines: []Node_routine{
				func(ctx context.Context, node TreeNode) {
					if node.Type() == NodeType_map_pair {
						switch node.Value().TypeKind() {
						case reflect.Int:
							node.Value().Set(node.Value().MustInt() * 10)
						}
					}
				},
			},
			expect: map[string]interface{}{
				"alice": map[string]interface{}{
					"age": 50,
				},
				"bob": map[string]interface{}{
					"firstname": "james",
				},
			},
		},
		{
			name: "替换覆盖map多种类型值",
			input: map[interface{}]interface{}{
				"misc": map[interface{}]interface{}{
					"int":    1,
					"int8":   int8(1),
					"int16":  int16(1),
					"int32":  int32(1),
					"int64":  int64(1),
					"uint":   uint(1),
					"uint8":  uint8(1),
					"uint16": uint16(1),
					"uint32": uint32(1),
					"uint64": uint64(1),
					// "float32": float32(3.14), // TODO，精度问题
					// "float64": float64(3.14), // TODO，精度问题
					"str": "str",
				},
			},
			jsonable: true,
			routines: []Node_routine{
				func(ctx context.Context, node TreeNode) {
					if node.Type() == NodeType_map_pair {
						val := node.Value()
						switch val.TypeKind() {
						case reflect.Int:
							val.Set(val.MustInt() + 1)
						case reflect.Int8:
							val.Set(val.MustInt8() + 1)
						case reflect.Int16:
							val.Set(val.MustInt16() + 1)
						case reflect.Int32:
							val.Set(val.MustInt32() + 1)
						case reflect.Int64:
							val.Set(val.MustInt64() + 1)
						case reflect.Uint:
							val.Set(val.MustUint() + 2)
						case reflect.Uint8:
							val.Set(val.MustUint8() + 2)
						case reflect.Uint16:
							val.Set(val.MustUint16() + 2)
						case reflect.Uint32:
							val.Set(val.MustUint32() + 2)
						case reflect.Uint64:
							val.Set(val.MustUint64() + 2)
						case reflect.Float32:
							val.Set(val.MustFloat32() + 1)
						case reflect.Float64:
							val.Set(val.MustFloat64() + 1)
						case reflect.String:
							val.Set(strings.Repeat(val.MustString(), 3))
						}
					}
				},
			},
			expect: map[string]interface{}{
				"misc": map[string]interface{}{
					"int":    2,
					"int8":   int8(2),
					"int16":  int16(2),
					"int32":  int32(2),
					"int64":  int64(2),
					"uint":   uint(3),
					"uint8":  uint8(3),
					"uint16": uint16(3),
					"uint32": uint32(3),
					"uint64": uint64(3),
					// "float32": float32(4.14),
					// "float64": float64(4.14),
					"str": "strstrstr",
				},
			},
		},

		{
			name: "删除map-key测试",
			input: map[interface{}]interface{}{
				"alice": map[interface{}]interface{}{
					"age": 5,
				},
				"bob": map[interface{}]interface{}{
					"firstname": "james",
					"lastname":  "bob",
				},
			},
			jsonable: true,
			routines: []Node_routine{
				func(ctx context.Context, node TreeNode) {
					// 删除lastname字段
					if node.Type() != NodeType_map_pair {
						return
					}
					s, err := node.Key().String()
					if err != nil {
						return
					}
					if s == "lastname" {
						node.Delete()
					}
				},
			},
			expect: map[string]interface{}{
				"alice": map[string]interface{}{
					"age": 5,
				},
				"bob": map[string]interface{}{
					"firstname": "james",
				},
			},
		},
		{
			name: "多routine测试（对mapkey执行：大写->特定删除->逆序）",
			input: map[interface{}]interface{}{
				"alice": map[interface{}]interface{}{
					"age": 5,
				},
				"bob": map[interface{}]interface{}{
					"firstname": "james",
					"lastname":  "bob",
				},
			},
			jsonable: true,
			routines: []Node_routine{
				func(ctx context.Context, node TreeNode) {
					// key修改为大写
					if node.Type() == NodeType_map_pair {
						s, err := node.Key().String()
						if err != nil {
							return
						}
						node.Key().Set(strings.ToUpper(s))
					}
				},

				func(ctx context.Context, node TreeNode) {
					// 删除key为LASTNAME的字段
					if node.Type() == NodeType_map_pair {
						s, err := node.Key().String()
						if err != nil {
							return
						}
						if s == "LASTNAME" {
							node.Delete()
						}
					}
				},

				func(ctx context.Context, node TreeNode) {
					// 逆序字符串
					if node.Type() == NodeType_map_pair {
						s, err := node.Key().String()
						if err != nil {
							t.Errorf("type assert fail")
							return
						}
						runes := []rune(s)
						for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
							runes[i], runes[j] = runes[j], runes[i]
						}
						node.Key().Set(string(runes))
					}
				},
			},
			expect: map[string]interface{}{
				"ECILA": map[string]interface{}{
					"EGA": 5,
				},
				"BOB": map[string]interface{}{
					"EMANTSRIF": "james",
				},
			},
		},
	}
	for _, v := range testCases {
		t.Run(v.name, func(t *testing.T) {

			tw := NewTreeWalker(
				WithRoutine(v.routines...),
				WithJsonableMap(),
			)
			got := tw.Walk(context.Background(), v.input)
			if !reflect.DeepEqual(got, v.expect) {
				t.Errorf("miss match: \n\tinput:  %+v\n\texpect:%+v\n\tgot:   %+v", v.input, v.expect, got)
			}
		})
	}
}

func Test_slice_routine(t *testing.T) {
	testCases := []struct {
		name     string
		enable   bool
		input    interface{}
		routines []Node_routine
		expect   interface{}
	}{
		{
			name: "修改slice成员测试",
			input: []interface{}{
				1,
				2,
				3,
			},
			expect: []interface{}{
				2,
				4,
				6,
			},
			routines: []Node_routine{
				func(ctx context.Context, node TreeNode) {
					if node.Type() != NodeType_slice_member {
						return
					}
					n := node.Value().MustInt()
					node.Value().Set(n << 1)
				},
			},
		},
		{
			name: "修改slice成员测试",
			input: []interface{}{
				1,
				nil,
				3,
			},
			expect: []interface{}{
				1,
				"hello",
				3,
			},
			routines: []Node_routine{
				func(ctx context.Context, node TreeNode) {
					if node.Type() != NodeType_slice_member {
						return
					}
					if node.Value().Interface() == nil {
						node.Value().Set("hello")
					}
				},
			},
		},
		{
			name: "删除slice成员测试 - 留下偶数",
			input: []interface{}{
				1,
				2,
				3,
				4,
				5,
			},
			expect: []interface{}{
				2,
				4,
			},
			routines: []Node_routine{
				func(ctx context.Context, node TreeNode) {
					if node.Type() != NodeType_slice_member {
						return
					}

					if node.Value().TypeKind() != reflect.Int {
						return
					}

					remainder := node.Value().MustInt() % 2
					if remainder != 0 {
						node.Delete()
					}
				},
			},
		},
		{
			name: "替换覆盖slice值测试",
			input: []interface{}{
				"string",
				5,
				nil,
			},
			routines: []Node_routine{
				func(ctx context.Context, node TreeNode) {
					// 修改替换slice中的值
					if node.Type() == NodeType_slice_member {
						switch node.Value().TypeKind() {
						case reflect.String:
							node.Value().Set(strings.Repeat(node.Value().MustString(), 4))
						case reflect.Int:
							node.Value().Set(node.Value().MustInt() * 100)
						}
					}
				},
			},
			expect: []interface{}{
				"stringstringstringstring",
				500,
				nil,
			},
		},
	}
	for _, v := range testCases {
		t.Run(v.name, func(t *testing.T) {

			tw := NewTreeWalker(
				WithRoutine(v.routines...),
			)
			got := tw.Walk(context.Background(), v.input)
			if !reflect.DeepEqual(got, v.expect) {
				t.Errorf("miss match: \n\tinput:  %+v\n\texpect:%+v\n\tgot:   %+v", v.input, v.expect, got)
			}
		})
	}
}

func Test_struct(t *testing.T) {
	type Tws_embed struct {
		EmbedPublicString  string `json:"embedPublicString"`
		EmbedPublicInt     int    `json:"embedPublicInt"`
		embedPrivateString string
		embedPrivateInt    int
	}

	type tws struct {
		PublicString  string `json:"publicString"`
		PublicInt     int    `json:"publicInt"`
		privateString string
		privateInt    int
		Tws_embed
	}

	testCases := []struct {
		name     string
		input    interface{}
		routines []Node_routine
		expect   interface{}
	}{
		{
			input:  tws{},
			expect: tws{},
		},
		{
			input:  &tws{},
			expect: &tws{},
		},
		{
			input: &tws{
				PublicString:  "foo",
				PublicInt:     1,
				privateString: "bar",
				privateInt:    2,
			},
			expect: &tws{
				PublicString:  "foofoofoofoofoo",
				PublicInt:     3,
				privateString: "bar",
				privateInt:    2,
			},
			routines: []Node_routine{
				func(ctx context.Context, node TreeNode) {
					if node.Type() != NodeType_struct_member {
						return
					}

					val := node.Value()
					switch val.TypeKind() {
					case reflect.String:
						val.Set(strings.Repeat(val.MustString(), 5))
					case reflect.Int:
						val.Set(val.MustInt() * 3)
					}
				},
			},
		},
		{
			input: &tws{
				PublicString: "foo",
				PublicInt:    1,
			},
			expect: &tws{
				PublicString: "foo",
				PublicInt:    1,
			},
			routines: []Node_routine{
				func(ctx context.Context, node TreeNode) {
					if node.Type() != NodeType_struct_member {
						return
					}
					node.Delete()
				},
			},
		},
	}
	for _, v := range testCases {
		t.Run(v.name, func(t *testing.T) {
			tw := NewTreeWalker(
				WithRoutine(v.routines...),
			)

			got := tw.Walk(context.Background(), v.input)
			vgot := reflect.ValueOf(got)
			if vgot.Kind() == reflect.Pointer || vgot.Kind() == reflect.Interface {
				got = vgot.Elem().Interface()
			}

			expect := v.expect
			vexpect := reflect.ValueOf(expect)
			if vexpect.Kind() == reflect.Pointer || vexpect.Kind() == reflect.Interface {
				expect = vexpect.Elem().Interface()
			}

			if !reflect.DeepEqual(got, expect) {
				t.Errorf("miss match: \n\tinput:  %+v\n\texpect:%+v\n\tgot:   %+v", v.input, v.expect, got)
			}
		})
	}
}

// // 最大递归深度测试
// func Test_MaxDepth(t *testing.T) {
// 	testCases := []struct {
// 		name     string
// 		input    interface{}
// 		jsonable bool
// 		routines []Node_routine
// 		expect   interface{}
// 		maxDepth int
// 	}{
// 		{
// 			name: "不限制深度测试",
// 			input: map[interface{}]interface{}{
// 				"a": map[interface{}]interface{}{
// 					"b": map[interface{}]interface{}{
// 						"c": map[interface{}]interface{}{
// 							"d": map[interface{}]interface{}{
// 								"e": map[interface{}]interface{}{
// 									"f": map[interface{}]interface{}{
// 										"g": []interface{}{
// 											"alice",
// 											"bob",
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			jsonable: true,
// 			routines: []Node_routine{
// 				func(ctx context.Context, node *TreeNode) Routine_action {
// 					if node.NType == NodeType_map_key || node.NType == NodeType_val_string {
// 						node.S = strings.ToUpper(node.String())
// 						return Routine_override
// 					}
// 					return Routine_blank
// 				},
// 			},
// 			expect: map[string]interface{}{
// 				"A": map[string]interface{}{
// 					"B": map[string]interface{}{
// 						"C": map[string]interface{}{
// 							"D": map[string]interface{}{
// 								"E": map[string]interface{}{
// 									"F": map[string]interface{}{
// 										"G": []interface{}{
// 											"ALICE",
// 											"BOB",
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			maxDepth: NoDepthLimit,
// 		},
// 		{
// 			name: "深度限制高于实际深度",
// 			input: map[interface{}]interface{}{
// 				"a": map[interface{}]interface{}{
// 					"b": map[interface{}]interface{}{
// 						"c": map[interface{}]interface{}{
// 							"d": map[interface{}]interface{}{
// 								"e": map[interface{}]interface{}{
// 									"f": map[interface{}]interface{}{
// 										"g": []interface{}{
// 											"alice",
// 											"bob",
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			jsonable: true,
// 			routines: []Node_routine{
// 				func(ctx context.Context, node *TreeNode) Routine_action {
// 					if node.NType == NodeType_map_key || node.NType == NodeType_val_string {
// 						node.S = strings.ToUpper(node.String())
// 						return Routine_override
// 					}
// 					return Routine_blank
// 				},
// 			},
// 			expect: map[string]interface{}{
// 				"A": map[string]interface{}{
// 					"B": map[string]interface{}{
// 						"C": map[string]interface{}{
// 							"D": map[string]interface{}{
// 								"E": map[string]interface{}{
// 									"F": map[string]interface{}{
// 										"G": []interface{}{
// 											"ALICE",
// 											"BOB",
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			maxDepth: 10,
// 		},
// 		{
// 			name: "深度限制低于实际深度",
// 			input: map[interface{}]interface{}{
// 				"a": map[interface{}]interface{}{
// 					"b": map[interface{}]interface{}{
// 						"c": map[interface{}]interface{}{
// 							"d": map[interface{}]interface{}{
// 								"e": map[interface{}]interface{}{
// 									"f": map[interface{}]interface{}{
// 										"g": []interface{}{
// 											"alice",
// 											"bob",
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			jsonable: true,
// 			routines: []Node_routine{
// 				func(ctx context.Context, node *TreeNode) Routine_action {
// 					if node.NType == NodeType_map_key || node.NType == NodeType_val_string {
// 						node.S = strings.ToUpper(node.String())
// 						return Routine_override
// 					}
// 					return Routine_blank
// 				},
// 			},
// 			expect: map[string]interface{}{ // 0
// 				"A": map[string]interface{}{ // 1
// 					"B": map[string]interface{}{ // 2
// 						"C": map[string]interface{}{ // 3
// 							"D": map[string]interface{}{ // 4
// 								"E": map[string]interface{}{ // 5
// 									"F": map[interface{}]interface{}{
// 										"g": []interface{}{
// 											"alice",
// 											"bob",
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			maxDepth: 5,
// 		},
// 		{
// 			name: "深度限制低于实际深度(两组map路径)",
// 			input: map[interface{}]interface{}{ // 0
// 				"a": map[interface{}]interface{}{ // 1
// 					"b": map[interface{}]interface{}{ // 2
// 						"c": map[interface{}]interface{}{ // 3
// 							"d": map[interface{}]interface{}{ // 4
// 								"e": map[interface{}]interface{}{
// 									"f": map[interface{}]interface{}{
// 										"g": []interface{}{
// 											"alice",
// 											"bob",
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 				"b": map[interface{}]interface{}{
// 					"c": map[interface{}]interface{}{
// 						"d": map[interface{}]interface{}{
// 							"e": map[interface{}]interface{}{
// 								"f": map[interface{}]interface{}{
// 									"g": []interface{}{
// 										"alice",
// 										"bob",
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			jsonable: true,
// 			routines: []Node_routine{
// 				func(ctx context.Context, node *TreeNode) Routine_action {
// 					if node.NType == NodeType_map_key || node.NType == NodeType_val_string {
// 						node.S = strings.ToUpper(node.String())
// 						return Routine_override
// 					}
// 					return Routine_blank
// 				},
// 			},
// 			expect: map[string]interface{}{ // 0
// 				"A": map[string]interface{}{ // 1
// 					"B": map[string]interface{}{ // 2
// 						"C": map[string]interface{}{ // 3
// 							"D": map[string]interface{}{ // 4
// 								"E": map[string]interface{}{ // 5
// 									"F": map[interface{}]interface{}{
// 										"g": []interface{}{
// 											"alice",
// 											"bob",
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 				"B": map[string]interface{}{ // 1
// 					"C": map[string]interface{}{ // 2
// 						"D": map[string]interface{}{ // 3
// 							"E": map[string]interface{}{ // 4
// 								"F": map[string]interface{}{ // 5
// 									"G": []interface{}{
// 										"alice",
// 										"bob",
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			maxDepth: 5,
// 		},
// 	}
// 	for _, v := range testCases {
// 		t.Run(v.name, func(t *testing.T) {
// 			tw := NewTreeWalker(
// 				WithMaxDepth(v.maxDepth),
// 				WithRoutine(v.routines...),
// 				WithJsonableMap(v.jsonable),
// 			)
// 			got := tw.Walk(context.Background(), v.input)
// 			if !reflect.DeepEqual(got, v.expect) {
// 				t.Errorf("miss match: \n\tinput:  %+v\n\texpect:%+v\n\tgot:   %+v", v.input, v.expect, got)
// 			}
// 		})
// 	}
// }

// type Person struct {
// 	Name   string
// 	Age    int
// 	Height float32
// }

// func Test_Struct(t *testing.T) {
// 	testCases := []struct {
// 		name    string
// 		input   interface{}
// 		routine []Node_routine
// 		expect  string
// 	}{
// 		{
// 			input: &Person{
// 				Name:   "alice",
// 				Age:    5,
// 				Height: 110.5,
// 			},
// 			routine: []Node_routine{
// 				func(ctx context.Context, node *TreeNode) Routine_action {
// 					if node.NType == NodeType_struct {
// 						p := node.Struct.(Person)
// 						p.Age *= 10
// 						p.Height += 58
// 						p.Name = strings.Repeat(p.Name, 3)
// 						node.Struct = p
// 						return Routine_override
// 					}
// 					return Routine_blank
// 				},
// 			},
// 		},
// 	}
// 	for _, v := range testCases {
// 		t.Run(v.name, func(t *testing.T) {
// 			tw := NewTreeWalker(
// 				WithRoutine(v.routine...),
// 			)
// 			got := tw.Walk(context.Background(), v.input)
// 			t.Logf("%+v", got)
// 		})
// 	}
// }

// func Benchmark_Walk(b *testing.B) {
// 	testCases := []struct {
// 		name   string
// 		input  string
// 		expect string
// 	}{
// 		{},
// 	}
// 	for _, v := range testCases {
// 		b.Run(v.name, func(b *testing.B) {
// 			for i := 0; i < b.N; i++ {
// 			}
// 		})
// 	}
// }

// type ABC struct {
// 	N int
// 	V string
// }

// func Test(t *testing.T) {
// 	testCases := []struct {
// 		name   string
// 		input  string
// 		expect string
// 	}{
// 		{},
// 	}
// 	for _, v := range testCases {
// 		t.Run(v.name, func(t *testing.T) {
// 			m := map[interface{}]string{1: "a", 2: "b"}
// 			b, e := json.Marshal(m)
// 			if e != nil {
// 				t.Errorf("error: %s", e.Error())
// 				return
// 			}
// 			t.Logf("%+s", string(b))
// 		})
// 	}
// }
