package etcd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/huangsc/blade/config"
)

// defaultParser 默认配置解析器
type defaultParser struct{}

// Parse 解析配置
func (p *defaultParser) Parse(data map[string]interface{}) (map[string]config.Value, error) {
	values := make(map[string]config.Value)
	for k, v := range data {
		values[k] = &defaultValue{value: v}
	}
	return values, nil
}

// defaultValue 默认配置值
type defaultValue struct {
	value interface{}
}

// Bool 获取布尔值
func (v *defaultValue) Bool() (bool, error) {
	switch val := v.value.(type) {
	case bool:
		return val, nil
	case string:
		return strconv.ParseBool(val)
	default:
		return false, config.ErrTypeAssert
	}
}

// Int 获取整数值
func (v *defaultValue) Int() (int64, error) {
	switch val := v.value.(type) {
	case int:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case float32:
		return int64(val), nil
	case float64:
		return int64(val), nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	default:
		return 0, config.ErrTypeAssert
	}
}

// Float 获取浮点值
func (v *defaultValue) Float() (float64, error) {
	switch val := v.value.(type) {
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, config.ErrTypeAssert
	}
}

// String 获取字符串值
func (v *defaultValue) String() (string, error) {
	switch val := v.value.(type) {
	case string:
		return val, nil
	case []byte:
		return string(val), nil
	case fmt.Stringer:
		return val.String(), nil
	default:
		return fmt.Sprintf("%v", val), nil
	}
}

// Duration 获取时间间隔
func (v *defaultValue) Duration() (time.Duration, error) {
	switch val := v.value.(type) {
	case time.Duration:
		return val, nil
	case int:
		return time.Duration(val), nil
	case int64:
		return time.Duration(val), nil
	case string:
		return time.ParseDuration(val)
	default:
		return 0, config.ErrTypeAssert
	}
}

// Time 获取时间值
func (v *defaultValue) Time() (time.Time, error) {
	switch val := v.value.(type) {
	case time.Time:
		return val, nil
	case string:
		return time.Parse(time.RFC3339, val)
	case int64:
		return time.Unix(val, 0), nil
	default:
		return time.Time{}, config.ErrTypeAssert
	}
}

// Slice 获取切片值
func (v *defaultValue) Slice() ([]config.Value, error) {
	if slice, ok := v.value.([]interface{}); ok {
		values := make([]config.Value, len(slice))
		for i, val := range slice {
			values[i] = &defaultValue{value: val}
		}
		return values, nil
	}
	return nil, config.ErrTypeAssert
}

// Map 获取映射值
func (v *defaultValue) Map() (map[string]config.Value, error) {
	if m, ok := v.value.(map[string]interface{}); ok {
		values := make(map[string]config.Value)
		for k, val := range m {
			values[k] = &defaultValue{value: val}
		}
		return values, nil
	}
	return nil, config.ErrTypeAssert
}

// Scan 将值扫描到结构体
func (v *defaultValue) Scan(dest interface{}) error {
	data, err := json.Marshal(v.value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
