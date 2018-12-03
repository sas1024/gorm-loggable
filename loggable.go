package loggable

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/gofrs/uuid"
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
	ID         uuid.UUID `gorm:"primary_key;"`
	CreatedAt  time.Time `sql:"DEFAULT:current_timestamp"`
	Action     string
	ObjectID   string      `gorm:"index"`
	ObjectType string      `gorm:"index"`
	RawObject  string      `sql:"type:JSON"`
	RawMeta    string      `sql:"type:JSON"`
	RawDiff    string      `sql:"type:JSON"`
	CreatedBy  string      `gorm:"index"`
	Object     interface{} `sql:"-"`
	Meta       interface{} `sql:"-"`
}

func (l *ChangeLog) prepareObject(objType reflect.Type) (err error) {
	obj := reflect.New(objType).Interface()
	err = json.Unmarshal([]byte(l.RawObject), obj)
	l.Object = obj
	return
}

func (l *ChangeLog) prepareMeta(objType reflect.Type) (err error) {
	obj := reflect.New(objType).Interface()
	err = json.Unmarshal([]byte(l.RawMeta), obj)
	l.Meta = obj
	return
}

func (l ChangeLog) Diff() (UpdateDiff, error) {
	var diff UpdateDiff
	err := json.Unmarshal([]byte(l.RawDiff), &diff)
	if err != nil {
		return nil, err
	}

	return diff, nil
}

func interfaceToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprint(v)
	}
}

func fetchChangeLogMeta(scope *gorm.Scope) []byte {
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

func isLoggable(value interface{}) bool {
	_, ok := value.(Interface)
	return ok
}

func isEnabled(value interface{}) bool {
	v, ok := value.(Interface)
	return ok && v.isEnabled()
}
