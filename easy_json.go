package EasyJSON

import (
	"encoding/json"
	"errors"
	"strings"
	"strconv"
	"reflect"
	"runtime"
	"fmt"
	"time"
)


/*
Go 的基本类型有

bool
string

int  int8  int16  int32  int64
uint uint8 uint16 uint32 uint64 uintptr

byte // uint8 的别名

rune // int32 的别名，表示一个 Unicode 码点

float32 float64

complex64 complex128

int, uint 和 uintptr 在 32 位系统上通常为 32 位宽，在 64 位系统上则为 64 位宽。
当你需要一个整数值时应使用 int 类型，除非你有特殊的理由使用固定大小或无符号的整数类型。


a是数组，s是切片
a := [3]int{6, 7, 8} // shorthand declaration to create array
s := []int{6, 7, 8} //creates and array and returns a slice reference

 */




type EasyJSON struct {
	jsonType int  // JSON类型: 对象或是数组

	// 底层的数据表示
	m map[string] interface{}
	a []interface{}
}

const (
	JSON_TYPE_INVALID = 0  // 0 -- 无效的JSON类型
	JSON_TYPE_OBJECT = 1   // 1 -- JSON对象
	JSON_TYPE_ARRAY = 2    // 2 -- JSON数组
)

// 错误列表定义

var (
	ErrInvalidJSONString = errors.New("invalid JSON string")
	ErrInvalidArguments = errors.New("invalid arguments")
	ErrIndexOutOfBounds = errors.New("index out of bounds")
	ErrFieldNotExists = errors.New("field not exists")
	ErrNotAnArray = errors.New("not an array")
	ErrNotAnObject = errors.New("not an object")
	ErrNotAStruct = errors.New("not a struct")
)


/*
从给定的jsonString解析出EasyJSON对象
返回
   如果成功，返回EasyJSON的指针,并且error为nil
   如果失败，EasyJSON的指针的指针为nil，error为具体的报错信息
 */
func Parse(jsonString string) (*EasyJSON, error) {
	jsonType := JSON_TYPE_INVALID

	// 根据第一个字符确定是JSON对象还是JSON数组
	for i := 0; i < len(jsonString); i++ {
		ch := jsonString[i]

		if ch == '{' {  // 确定是JSON对象
			jsonType = JSON_TYPE_OBJECT
			break
		} else if ch == '[' {  // 确定是JSON数组
			jsonType = JSON_TYPE_ARRAY
			break
		}
	}

	if jsonType == JSON_TYPE_INVALID {
		return nil, ErrInvalidJSONString
	}

	var data interface{}

	err := json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return nil, err
	}

	// slog("data[%v]", data)

	if jsonType == JSON_TYPE_OBJECT {
		return &EasyJSON{jsonType, data.(map[string] interface{}), nil}, nil
	} else {
		return &EasyJSON{jsonType, nil, data.([]interface{})}, nil
	}
}

/**
生成一个JSON对象
第0个参数为name，第1个参数为value，第2个参数为name，第3个参数为value，... 依此类推
name必须为string类型，value可以为任意类型
 */
func Object(args ...interface{}) *EasyJSON {
	argsCount := len(args)

	// 参数个数必须为偶数个
	if argsCount % 2 != 0 {
		return nil
	}

	m := make(map[string] interface{})
	name := ""

	for index, value := range args {
		if index % 2 == 0 {  // name
			name = value.(string)
		} else {  // value
			m[name] = valueEncoder(value)
		}
	}

	return &EasyJSON{JSON_TYPE_OBJECT, m, nil}
}

/**
生成一个JSON数组
参数可以是任意类型
 */
func Array(args ...interface{}) *EasyJSON {
	var a []interface{}
	for _, arg := range args {
		a = append(a, valueEncoder(arg))
	}
	// slog("a[%v]", a)
	return &EasyJSON{JSON_TYPE_ARRAY, nil, a}
}



/**
获取EasyJSON的类型
返回:
    1 -- JSON_TYPE_OBJECT JSON对象
    2 -- JSON_TYPE_ARRAY  JSON数组
 */
func (easyJSON *EasyJSON) GetJSONType() int {
	return easyJSON.jsonType
}


/**
获取底层的JSON数据表示
 */
func (easyJSON *EasyJSON) GetData() interface{}  {
	jsonType := easyJSON.GetJSONType()
	if jsonType == JSON_TYPE_OBJECT {
		return easyJSON.m
	} else {
		return easyJSON.a
	}
}

func (easyJSON *EasyJSON) Get(path string) (interface{}, error)  {
	nameList := parsePath(path)
	var value interface{}

	jsonType := easyJSON.GetJSONType()
	if jsonType == JSON_TYPE_OBJECT {
		value = easyJSON.m
	} else {
		value = easyJSON.a
	}


	for _, name := range nameList {
		ch := name[0]
		if ch == '[' {  // 表明是数组
			size := len(name)
			name = name[1 : size - 1]  // 去除前后中括号
			index, _ := strconv.Atoi(name)

			a := value.([]interface{})
			if index >= len(a) {  // 数组越界
				return nil, ErrIndexOutOfBounds
			}
			value = a[index]
		} else { // 表明是对象
			m := value.(map[string] interface{})
			val, ok := m[name]
			if !ok {
				return nil, ErrFieldNotExists
			}
			value = val
		}
	}
	return value, nil
}


func (easyJSON *EasyJSON) Opt(path string, defaultValue interface{}) interface{} {
	value, err := easyJSON.Get(path)
	if err == nil {
		return value
	}

	return defaultValue
}


func (easyJSON *EasyJSON) Exists(path string) bool {
	_, err := easyJSON.Get(path)
	return err == nil
}


func (easyJSON *EasyJSON) GetInt64(path string) (int64, error) {
	value, err := easyJSON.Get(path)
	if err != nil {
		return 0, err
	}

	switch value.(type) {
	case int:
		return int64(value.(int)), nil
	case int8:
		return int64(value.(int8)), nil
	case int16:
		return int64(value.(int16)), nil
	case int32:
		return int64(value.(int32)), nil
	case int64:
		return value.(int64), nil
	case uint:
		return int64(value.(uint)), nil
	case uint8:
		return int64(value.(uint8)), nil
	case uint16:
		return int64(value.(uint16)), nil
	case uint32:
		return int64(value.(uint32)), nil
	case uint64:
		return int64(value.(uint64)), nil
	case float32:
		return int64(value.(float32)), nil
	case float64:
		return int64(value.(float64)), nil
	}

	// 理论上到不了这里
	return 0, nil
}

func (easyJSON *EasyJSON) OptInt64(path string, defaultValue int64) int64 {
	value, err := easyJSON.GetInt64(path)
	if err == nil {
		return value
	}

	return defaultValue
}

func (easyJSON *EasyJSON) GetFloat64(path string) (float64, error) {
	value, err := easyJSON.Get(path)
	if err != nil {
		return 0, err
	}

	switch value.(type) {
	case float64:
		return value.(float64), nil
	case float32:
		return float64(value.(float32)), nil
	}

	// 理论上到不了这里
	return 0, nil
}


func (easyJSON *EasyJSON) OptFloat64(path string, defaultValue float64) float64 {
	value, err := easyJSON.GetFloat64(path)
	if err == nil {
		return value
	}

	return defaultValue
}

func (easyJSON *EasyJSON) GetBoolean(path string) (bool, error) {
	value, err := easyJSON.Get(path)
	if err != nil {
		return false, err
	}

	return value.(bool), nil
}

func (easyJSON *EasyJSON) OptBoolean(path string, defaultValue bool) bool {
	value, err := easyJSON.GetBoolean(path)
	if err == nil {
		return value
	}

	return defaultValue
}


func (easyJSON *EasyJSON) GetString(path string) (string, error) {
	value, err := easyJSON.Get(path)
	if err != nil {
		return "", err
	}

	return value.(string), nil
}

func (easyJSON *EasyJSON) OptString(path string, defaultValue string) string {
	value, err := easyJSON.GetString(path)
	if err == nil {
		return value
	}

	return defaultValue
}

func (easyJSON *EasyJSON) GetObject(path string) (*EasyJSON, error) {
	value, err := easyJSON.Get(path)
	if err != nil {
		return nil, err
	}

	return &EasyJSON{JSON_TYPE_OBJECT, value.(map[string] interface{}), nil}, nil
}


func (easyJSON *EasyJSON) OptObject(path string, defaultValue *EasyJSON) *EasyJSON {
	value, err := easyJSON.GetObject(path)
	if err == nil {
		return value
	}

	return defaultValue
}


func (easyJSON *EasyJSON) GetArray(path string) (*EasyJSON, error) {
	value, err := easyJSON.Get(path)
	if err != nil {
		return nil, err
	}

	return &EasyJSON{JSON_TYPE_ARRAY, nil, value.([]interface{})}, nil
}

func (easyJSON *EasyJSON) OptArray(path string, defaultValue *EasyJSON) *EasyJSON {
	value, err := easyJSON.GetArray(path)
	if err == nil {
		return value
	}

	return defaultValue
}

func (easyJSON *EasyJSON) Set(path string, value interface{}) error  {
	value = valueEncoder(value)

	nameList := parsePath(path)
	var iter interface{}

	jsonType := easyJSON.GetJSONType()
	if jsonType == JSON_TYPE_OBJECT {
		iter = easyJSON.m
	} else {
		iter = easyJSON.a
	}

	nameCount := len(nameList)
	// slog("nameCount[%v]", nameCount)

	counter := 0
	for _, name := range nameList {
		counter++
		ch := name[0]
		if ch == '[' {  // 表明是数组
			size := len(name)
			name = name[1 : size - 1]  // 去除前后中括号
			index, _ := strconv.Atoi(name)

			a := iter.([]interface{})
			if index >= len(a) {  // 数组越界
				return ErrIndexOutOfBounds
			}

			if counter == nameCount {  // 已经到达path的终点
				a[index] = value
				return nil
			}

			// 还需要继续遍历
			iter = a[index]
		} else { // 表明是对象
			m := iter.(map[string] interface{})
			if counter == nameCount {  // 已经到达path的终点
				m[name] = value
				return nil
			}

			// 还需要继续遍历
			val, ok := m[name]
			if !ok {
				return ErrFieldNotExists
			}
			iter = val
		}
	}
	return nil
}



func (easyJSON *EasyJSON) Append(path string, value interface{}) error {
	value = valueEncoder(value)

	var iter interface{}

	jsonType := easyJSON.GetJSONType()
	if jsonType == JSON_TYPE_OBJECT {
		iter = easyJSON.m
	} else {
		iter = easyJSON.a
	}

	// 如果path为空字符串，表示在最外层进行Append操作
	if path == "" {
		if jsonType != JSON_TYPE_ARRAY {
			return ErrNotAnArray
		}

		easyJSON.a = append(easyJSON.a, value)
		return nil
	}

	nameList := parsePath(path)
	nameCount := len(nameList)
	// slog("nameCount[%v]", nameCount)

	counter := 0
	for _, name := range nameList {
		counter++
		ch := name[0]
		if ch == '[' {
			size := len(name)
			name = name[1 : size - 1]  // 去除前后中括号
			index, _ := strconv.Atoi(name)

			a := iter.([]interface{})
			if index >= len(a) {  // 数组越界
				return ErrIndexOutOfBounds
			}

			if counter == nameCount {  // 已经到达path的终点
				a[index] = append(a[index].([]interface{}), value)
				return nil
			}

			// 还需要继续遍历
			iter = a[index]
		} else {
			m := iter.(map[string] interface{})
			if counter == nameCount {  // 已经到达path的终点
				m[name] = append(m[name].([]interface{}), value)
				return nil
			}

			// 还需要继续遍历
			val, ok := m[name]
			if !ok {
				return ErrFieldNotExists
			}
			iter = val
		}
	}
	return nil
}


/*
返回JSON字符串
因为
var arr []interface{}
json.Marshal(arr) 会返回null
所以要自定义String()方法

json/encode.go源码中的注释
Array and slice values encode as JSON arrays, except that
[]byte encodes as a base64-encoded string, and a nil slice
encodes as the null JSON value.
 */
func (easyJSON *EasyJSON) String() string {
	jsonType := easyJSON.GetJSONType()
	if jsonType == JSON_TYPE_OBJECT {
		return toJSONString(easyJSON.m, JSON_TYPE_OBJECT)
	} else {
		return toJSONString(easyJSON.a, JSON_TYPE_ARRAY)
	}
}

func toJSONString(json interface{}, jsonType int) string {
	jsonString := ""
	first := true

	if jsonType == JSON_TYPE_OBJECT {
		jsonString += "{"
		for k, v := range json.(map[string] interface{}) {
			// slog("k[%v], v[%v]", k, v)

			if !first {
				jsonString += ","
			}
			first = false

			jsonString += `"` + k + `":`

			if v == nil {
				jsonString += "null"
			} else {
				kind := reflect.TypeOf(v).Kind().String()
				// slog("kind[%s]", kind)


				if kind == "string" {
					jsonString += `"` + v.(string) + `"`
				} else if kind == "map" {
					jsonString += toJSONString(v, JSON_TYPE_OBJECT)
				} else if kind == "slice" {
					jsonString += toJSONString(v, JSON_TYPE_ARRAY)
				} else {
					jsonString += fmt.Sprintf("%v", v)
				}
			}
		}
		jsonString += "}"
	} else {
		jsonString += "["
		for _, v := range json.([] interface{}) {

			if !first {
				jsonString += ","
			}
			first = false

			if v == nil {
				jsonString += "null"
			} else {
				kind := reflect.TypeOf(v).Kind().String()
				// slog("kind[%s]", kind)


				if kind == "string" {
					jsonString += `"` + v.(string) + `"`
				} else if kind == "map" {
					jsonString += toJSONString(v, JSON_TYPE_OBJECT)
				} else if kind == "slice" {
					jsonString += toJSONString(v, JSON_TYPE_ARRAY)
				} else {
					jsonString += fmt.Sprintf("%v", v)
				}
			}
		}
		jsonString += "]"
	}

	return jsonString
}

func valueEncoder(val interface{}) interface{}  {
	if val == nil {
		return nil
	}

	t := reflect.TypeOf(val)
	k := t.Kind()

	// 如果是EasyJSON类型，获取其底层的数据
	if strings.Contains(t.String(), "EasyJSON") {
		json := val.(*EasyJSON)
		if json.GetJSONType() == JSON_TYPE_OBJECT {
			return json.m
		} else {
			return json.a
		}
	}

	// 如果是指针，取其指向的值
	if k == reflect.Ptr {
		v := reflect.ValueOf(val)
		val = v.Elem().Interface()
		k = reflect.TypeOf(val).Kind()
	}

	if k == reflect.Struct {
		return structEncoder(val)
	}

	if k == reflect.Array || k == reflect.Slice {
		return arrayEncoder(val)
	}

	return val
}

func arrayEncoder(val interface{}) []interface{} {
	a := []interface{}{}

	v := reflect.ValueOf(val)
	n := v.Len()
	for i := 0; i < n; i++ {
		// slog("i = %d, elem = %v", i, v.Index(i))
		elem := valueEncoder(v.Index(i).Interface())
		a = append(a, elem)
	}

	return a
}

func structEncoder(val interface{}) map[string] interface{} {
	m := map[string] interface{}{}

	t := reflect.TypeOf(val)
	v := reflect.ValueOf(val)
	// k := t.Kind()
	// slog("kind[%v]", k)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		// fieldType := field.Type
		// slog("fieldName[%s], fieldType[%v]", fieldName, fieldType)
		fieldValue := valueEncoder(v.Field(i).Interface())
		fieldTag := field.Tag.Get("json")  // 取json的Tag

		// slog("fieldValue[%v], fieldTag[%s]", fieldValue, fieldTag)

		if len(fieldTag) > 0 {  // 优先用fieldTag
			m[fieldTag] = fieldValue
		} else {
			m[fieldName] = fieldValue
		}
	}

	return m
}


/*
分析路径，返回路径切片
 */
func parsePath(path string) []string {
	path = strings.TrimSpace(path)
	var nameList []string

	snippets := strings.Split(path, ".")
	for _, snippet := range snippets {
		size := len(snippet)
		i := 0
		j := 0
		var name string
		beginChar := snippet[0]

		for {
			ch := snippet[j]
			if beginChar == '[' { // 是数组索引
				if ch == ']' {
					j++
					name = snippet[i : j]
					nameList = append(nameList, name)
					i = j
					if i >= size {
						break
					}
					beginChar = snippet[i]
				}
			} else {  // 是对象字段
				if j + 1 == size {
					name = snippet[i : j + 1]
					nameList = append(nameList, name)
					break
				}
				if ch == '[' {
					name = snippet[i : j]
					nameList = append(nameList, name)
					i = j

					if i >= size {
						break
					}
					beginChar = snippet[i]
				}
			}
			j++
		}
	}

	return nameList
}

/*
判断是否为基本类型
 */
func isPrimitiveType(v interface{}) bool {
	// byte会返回uint8
	// rune会返回int32

	var typeOf = reflect.TypeOf(v).String()
	return typeOf == "bool" ||
			typeOf == "string" ||
			typeOf == "int" ||
			typeOf == "int8" ||
			typeOf == "int16" ||
			typeOf == "int32" ||
			typeOf == "int64" ||
			typeOf == "uint" ||
			typeOf == "uint8" ||
			typeOf == "uint16" ||
			typeOf == "uint32" ||
			typeOf == "uint64" ||
			typeOf == "uintptr" ||
			typeOf == "float32" ||
			typeOf == "float64" ||
			typeOf == "complex64" ||
			typeOf == "complex128"

}

func slog(format string, args ...interface{}) {
	// 获取时间
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	_, file, line, _ := runtime.Caller(1)
	// 变长参数，参考：  http://www.cnblogs.com/sysnap/p/6860671.html
	content := fmt.Sprintf(format, args...)

	fmt.Printf("[%s][%s][%05d]%s\n", timestamp, file, line, content)
}