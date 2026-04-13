package typecast

import (
	"fmt"
	"strconv"
	"time"
)

// Caster converts raw WAL column values to typed Go values.
type Caster struct {
	timeFmt string
}

// New returns a Caster. If timeFmt is empty, time.RFC3339 is used.
func New(timeFmt string) *Caster {
	if timeFmt == "" {
		timeFmt = time.RFC3339
	}
	return &Caster{timeFmt: timeFmt}
}

// ToString coerces v to a string.
func (c *Caster) ToString(v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}
	switch val := v.(type) {
	case string:
		return val, nil
	case []byte:
		return string(val), nil
	case int64:
		return strconv.FormatInt(val, 10), nil
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(val), nil
	case time.Time:
		return val.Format(c.timeFmt), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// ToInt64 coerces v to int64.
func (c *Caster) ToInt64(v interface{}) (int64, error) {
	if v == nil {
		return 0, nil
	}
	switch val := v.(type) {
	case int64:
		return val, nil
	case float64:
		return int64(val), nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("typecast: cannot convert %T to int64", v)
	}
}

// ToBool coerces v to bool.
func (c *Caster) ToBool(v interface{}) (bool, error) {
	if v == nil {
		return false, nil
	}
	switch val := v.(type) {
	case bool:
		return val, nil
	case int64:
		return val != 0, nil
	case float64:
		return val != 0, nil
	case string:
		return strconv.ParseBool(val)
	default:
		return false, fmt.Errorf("typecast: cannot convert %T to bool", v)
	}
}
