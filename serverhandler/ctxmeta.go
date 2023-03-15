package serverhandler

import (
	"errors"
	"sync"
)

type MateData struct {
	lock sync.RWMutex
	data map[interface{}]interface{}
}

func (m *MateData) SetInterface(key, val interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.data[key] = val
}

func (m *MateData) GetString(key interface{}) string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.data[key].(string)
}

func (m *MateData) GetInt64(key interface{}) int64 {
	m.lock.RLock()
	defer m.lock.RUnlock()

	val := m.data[key]
	switch val.(type) {
	case int64:
		return val.(int64)
	case int:
		return int64(val.(int))
	case int8:
		return int64(val.(int8))
	case int16:
		return int64(val.(int16))
	case int32:
		return int64(val.(int32))
	}
	panic(errors.New("non convertable type 'int64'"))
}

func (m *MateData) GetInt32(key interface{}) int32 {
	m.lock.RLock()
	defer m.lock.RUnlock()

	val := m.data[key]
	switch val.(type) {
	case int32:
		return val.(int32)
	case int:
		return int32(val.(int))
	case int8:
		return int32(val.(int8))
	case int16:
		return int32(val.(int16))
	}
	panic(errors.New("non convertable type 'int32'"))
}

func (m *MateData) GetUint32(key interface{}) uint32 {
	m.lock.RLock()
	defer m.lock.RUnlock()

	val := m.data[key]
	switch val.(type) {
	case uint32:
		return val.(uint32)
	case uint:
		return uint32(val.(uint))
	case uint8:
		return uint32(val.(uint8))
	case uint16:
		return uint32(val.(uint16))
	}
	panic(errors.New("non convertable type 'uint32'"))
}

func (m *MateData) GetUint64(key interface{}) uint64 {
	m.lock.RLock()
	defer m.lock.RUnlock()

	val := m.data[key]
	switch val.(type) {
	case uint64:
		return val.(uint64)
	case uint:
		return uint64(val.(uint))
	case uint8:
		return uint64(val.(uint8))
	case uint16:
		return uint64(val.(uint16))
	case uint32:
		return uint64(val.(uint32))
	}
	panic(errors.New("non convertable type 'uint64'"))
}

func (m *MateData) GetInterface(key interface{}) interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.data[key]
}
