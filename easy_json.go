package EasyJSON

import (
	"encoding/json"
	"errors"
	"strings"
	"strconv"
	"reflect"
)

type EasyJSON struct {
	jsonType int  // JSON类型: 对象或是数组
	data interface{}  // 底层的数据表示
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

	return &EasyJSON{jsonType, data}, nil
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
			if strings.Contains(reflect.TypeOf(value).String(), "EasyJSON") {  // 如果是EasyJSON类型，获取其底层的数据
				value = value.(*EasyJSON).data
			}
			m[name] = value
		}
	}

	return &EasyJSON{JSON_TYPE_OBJECT, m}
}


/**
生成一个JSON数组
参数可以是任意类型
 */
func Array(args ...interface{}) *EasyJSON {
	var a []interface{}
	for _, arg := range args {
		if strings.Contains(reflect.TypeOf(arg).String(), "EasyJSON") {  // 如果是EasyJSON类型，获取其底层的数据
			a = append(a, arg.(*EasyJSON).data)
		} else {
			a = append(a, arg)
		}
	}
	return &EasyJSON{JSON_TYPE_ARRAY, a}
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
	return easyJSON.data
}

func (easyJSON *EasyJSON) Get(path string) (interface{}, error)  {
	nameList := parsePath(path)
	var value interface{}

	value = easyJSON.data

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
	case float64:
		return int64(value.(float64)), nil
	case float32:
		return int64(value.(float32)), nil
	case int64:
		return value.(int64), nil
	case int:
		return int64(value.(int)), nil
	}

	// 理论上到不了这里
	return 0, nil
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

func (easyJSON *EasyJSON) GetBoolean(path string) (bool, error) {
	value, err := easyJSON.Get(path)
	if err != nil {
		return false, err
	}

	return value.(bool), nil
}

func (easyJSON *EasyJSON) GetString(path string) (string, error) {
	value, err := easyJSON.Get(path)
	if err != nil {
		return "", err
	}

	return value.(string), nil
}

func (easyJSON *EasyJSON) GetObject(path string) (*EasyJSON, error) {
	value, err := easyJSON.Get(path)
	if err != nil {
		return nil, err
	}

	return &EasyJSON{JSON_TYPE_OBJECT, value}, nil
}


func (easyJSON *EasyJSON) GetArray(path string) (*EasyJSON, error) {
	value, err := easyJSON.Get(path)
	if err != nil {
		return nil, err
	}

	return &EasyJSON{JSON_TYPE_ARRAY, value}, nil
}




/*
返回JSON字符串
 */
func (easyJSON *EasyJSON) String() string {
	b, err := json.Marshal(easyJSON.data)
	if err == nil {
		return string(b)
	}
	return ""
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

