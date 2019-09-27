package loggable

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/jinzhu/copier"
	"reflect"
)

type identityMap map[string]interface{}

// identityManager is used as cache.
type identityManager struct {
	m identityMap
}

func newIdentityManager() *identityManager {
	return &identityManager{
		m: make(identityMap),
	}
}

func (im *identityManager) save(value, pk interface{}) {
	t := reflect.TypeOf(value)
	newValue := reflect.New(t).Interface()
	err := copier.Copy(&newValue, value)
	if err != nil {
		panic(err)
	}

	im.m[genIdentityKey(t, pk)] = newValue
}

func (im identityManager) get(value, pk interface{}) interface{} {
	t := reflect.TypeOf(value)
	key := genIdentityKey(t, pk)
	m, ok := im.m[key]
	if !ok {
		return nil
	}

	return m
}

func genIdentityKey(t reflect.Type, pk interface{}) string {
	key := fmt.Sprintf("%v_%s", pk, t.Name())
	b := md5.Sum([]byte(key))

	return hex.EncodeToString(b[:])
}
