package loggable

import "reflect"

type Option func(options *options)

type options struct {
	lazyUpdate       bool
	lazyUpdateFields []string
	metaTypes        map[string]reflect.Type
	objectTypes      map[string]reflect.Type
}

func LazyUpdate(fields ...string) Option {
	return func(options *options) {
		options.lazyUpdate = true
		options.lazyUpdateFields = fields
	}
}

func RegObjectType(objectType string, objectStruct interface{}) Option {
	return func(options *options) {
		options.objectTypes[objectType] = reflect.Indirect(reflect.ValueOf(objectStruct)).Type()
	}
}

func RegMetaType(objectType string, metaType interface{}) Option {
	return func(options *options) {
		options.metaTypes[objectType] = reflect.Indirect(reflect.ValueOf(metaType)).Type()
	}
}
