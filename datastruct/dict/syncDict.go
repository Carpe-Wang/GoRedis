package dict

import "sync"

// SyncDict wraps a map, it is not thread safe
type SyncDict struct {
	m sync.Map
}

// MakeSyncDict makes a new map
func MakeSyncDict() *SyncDict {
	return &SyncDict{}
}

// Get returns the binding value and whether the key is exist
func (dict *SyncDict) Get(key string) (val interface{}, exists bool) {
	val, ok := dict.m.Load(key)
	return val, ok
}

// Len returns the number of dict
func (dict *SyncDict) Len() int {
	lenth := 0
	dict.m.Range(func(k, v interface{}) bool {
		lenth++
		return true
	})
	return lenth
}

// Put puts key value into dict and returns the number of new inserted key-value
func (dict *SyncDict) Put(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	dict.m.Store(key, val)
	if existed {
		return 0
	}
	return 1
}

// PutIfAbsent puts value if the key is not exists and returns the number of updated key-value
func (dict *SyncDict) PutIfAbsent(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	if existed {
		return 0
	}
	dict.m.Store(key, val)
	return 1
}

// PutIfExists puts value if the key is exist and returns the number of inserted key-value
func (dict *SyncDict) PutIfExists(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	if existed {
		dict.m.Store(key, val)
		return 1
	}
	return 0
}

// Remove removes the key and return the number of deleted key-value
func (dict *SyncDict) Remove(key string) (result int) {
	_, existed := dict.m.Load(key)
	dict.m.Delete(key)
	if existed {
		return 1
	}
	return 0
}

// Keys returns all keys in dict
func (dict *SyncDict) Keys() []string {
	result := make([]string, dict.Len())
	i := 0
	dict.m.Range(func(key, value interface{}) bool {
		result[i] = key.(string)
		i++
		return true
	})
	return result
}

// ForEach traversal the dict
func (dict *SyncDict) ForEach(consumer Consumer) {
	dict.m.Range(func(key, value interface{}) bool {
		consumer(key.(string), value)
		return true
	})
}

// RandomKeys randomly returns keys of the given number, may contain duplicated key
func (dict *SyncDict) RandomKeys(limit int) []string {
	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		dict.m.Range(func(key, value interface{}) bool {
			result[i] = key.(string)
			return false
		})
	}
	return result

}

// RandomDistinctKeys randomly returns keys of the given number, won't contain duplicated key
func (dict *SyncDict) RandomDistinctKeys(limit int) []string {
	result := make([]string, limit)
	i := 0
	dict.m.Range(func(key, value interface{}) bool {
		result[i] = key.(string)
		i++
		if i == limit {
			return false
		}
		return true
	})
	return result
}

// Clear removes all keys in dict
func (dict *SyncDict) Clear() {
	*dict = *MakeSyncDict()
}
