// Package config provides three ways to set global variables.
// 1. From system environment
// 2. From json file
// 3. From default value(hard code)
// And in strict ORDER.
//
// When define a config structure, use tag 'json', 'env' to specify the keys, 'def' to set default value.
// The tag value '-' will omit the item
//
// Only 'all int', 'bool', 'string' fully supported.
// 支持的格式：各种int、bool和string
//
// [NOTICE] String should be ALWAYS provided DEFAULT value!!!
// [高能预警] String类型最好每个字段都设置默认值。
//
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
)

// SharedLeaf definition
type SharedLeaf struct {
	ListenIP   string `json:"ListenIP"   env:"LISTEN_IP"   def:"0.0.0.0"` // 监听地址
	ListenPort int    `json:"ListenPort" env:"LISTEN_PORT" def:"0"`       // 监听端口

	// skeleton conf
	GoLen              int `def:"8000"`
	TimerDispatcherLen int `def:"1000"`
	ChanRPCLen         int `def:"8000"`
	MaxConnNum         int `def:"5000"`

	// gate conf
	PendingWriteNum int    `def:"1000"`
	LenMsgLen       int    `def:"2"` //消息头长度2个字节
	MinMsgLen       int    `def:"2"`
	MaxMsgLen       uint32 `def:"65535"`

	// other
	SlowOpThresholdMs int64 `def:"20"`   // 记录慢命令
	SlowResponseMs    int64 `def:"2000"` //记录慢响应时间
}

// SharedBeego definition
type SharedBeego struct {
	HTTPPort int    `json:"HTTPPort" env:"HTTP_PORT" def:"8080"`
	AppName  string `json:"AppName"  env:"APP_NAME"  def:"my_game"`
	RunMode  string `json:"RunMode"  env:"RUN_MODE"  def:"dev"`
}

// Parse 来源优先级: environment > json > default
func Parse(filepath string, st interface{}) {
	// From json
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(data, &st)
	if err != nil {
		fmt.Println(err)
	}

	setFromEnvOrDefault(st)
}

func setFromEnvOrDefault(st interface{}) {
	// From environment or set as default
	values := reflect.ValueOf(st).Elem()
	types := values.Type()
	fieldNum := types.NumField()

	for i := 0; i < fieldNum; i++ {
		t := types.Field(i)
		v := values.Field(i)

		if v.CanSet() == false {
			panic(fmt.Sprintf("[Config Error]%s Field %s Cannot set.", types.Name(), t.Name))
		}

		if t.Name == "SharedLeaf" {
			x := v.Interface().(SharedLeaf)
			setFromEnvOrDefault(&x)
			v.Set(reflect.ValueOf(x))
			continue
		} else if t.Name == "SharedBeego" {
			x := v.Interface().(SharedBeego)
			setFromEnvOrDefault(&x)
			v.Set(reflect.ValueOf(x))
			continue
		}

		// Set from environment
		if envKey := t.Tag.Get("env"); envKey != "" && envKey != "-" {
			if envVal := os.Getenv(envKey); envVal != "" {
				setValue(&v, envVal)
			}
		}

		// 如果设置了环境变量，或已经有值，忽略默认值
		if isSet(&v) {
			continue
		}

		// Set as default
		if def := t.Tag.Get("def"); def != "" && def != "-" {
			setValue(&v, def)
		}
	}
}

// 检查field是否有值
// 约定空字符串("")、0和false为未设置初始值
func isSet(field *reflect.Value) bool {
	switch field.Kind() {
	case reflect.String:
		return len(field.Interface().(string)) > 0
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		return field.Int() > 0
	case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
		return field.Uint() > 0
	case reflect.Bool:
		return field.Interface().(bool)
	}

	return true
}

// 将strVal转换成对应的类型并赋值
func setValue(field *reflect.Value, strVal string) {
	t := field.Type()
	switch field.Kind() {
	case reflect.String:
		field.SetString(strVal)
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			fmt.Printf("[Config Error]Invalid value: %s(%T), got %v\n", t.Name(), field.Interface(), strVal)
		} else {
			field.SetInt(intVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
		intVal, err := strconv.ParseUint(strVal, 10, 64)
		if err != nil {
			fmt.Printf("[Config Error]Invalid value: %s(%T), got %v\n", t.Name(), field.Interface(), strVal)
		} else {
			field.SetUint(intVal)
		}
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(strVal)
		if err != nil {
			fmt.Printf("[Config Error]Invalid value: %s(%T), got %v\n", t.Name(), field.Interface(), strVal)
		} else {
			field.SetBool(boolVal)
		}
	}
}
