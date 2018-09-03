package loggable

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/jinzhu/gorm"
)

// Interface is used to get metadata from your models.
type Interface interface {
	// Meta should return structure, that can be converted to json.
	Meta() interface{}
	// lock makes available only embedding structures.
	lock()
	// check if callback enabled
	isEnabled() bool
	// enable/disable loggable
	Enable(v bool)
}

// LoggableModel is a root structure, which implement Interface.
// Embed LoggableModel to your model so that Plugin starts tracking changes.
type LoggableModel struct {
	Disabled bool `sql:"-"`
}

func (LoggableModel) Meta() interface{} { return nil }
func (LoggableModel) lock()             {}
func (l LoggableModel) isEnabled() bool { return !l.Disabled }
func (l LoggableModel) Enable(v bool)   { l.Disabled = !v }

// ChangeLog is a main entity, which used to log changes.
type ChangeLog struct {
	ID         string    `gorm:"type:uuid;primary_key;"`
	CreatedAt  time.Time `sql:"DEFAULT:current_timestamp"`
	Action     string
	ObjectID   string      `gorm:"index"`
	ObjectType string      `gorm:"index"`
	RawObject  JSONB       `sql:"type:JSONB"`
	RawMeta    JSONB       `sql:"type:JSONB"`
	Object     interface{} `sql:"-"`
	Meta       interface{} `sql:"-"`
}

func (l *ChangeLog) prepareObject(objType reflect.Type) (err error) {
	l.Object, err = l.RawObject.unmarshal(objType)
	return
}

func (l *ChangeLog) prepareMeta(objType reflect.Type) (err error) {
	l.Meta, err = l.RawMeta.unmarshal(objType)
	return
}

func interfaceToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprint(v)
	}
}

func fetchChangeLogMeta(scope *gorm.Scope) JSONB {
	val, ok := scope.Value.(Interface)
	if !ok {
		return nil
	}
	data, err := json.Marshal(val.Meta())
	if err != nil {
		panic(err)
	}
	return data
}

func isLoggable(scope *gorm.Scope) bool {
	_, ok := scope.Value.(Interface)
	return ok
}

func isEnabled(scope *gorm.Scope) bool {
	v, ok := scope.Value.(Interface)
	return ok && v.isEnabled()
}
