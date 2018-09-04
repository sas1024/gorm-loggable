package loggable

import (
	"encoding/json"
	"reflect"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
)

type LoggablePlugin interface {
	// Deprecated: Use SetUserAndWhere instead.
	SetUser(user string) *gorm.DB
	// Deprecated: Use SetUserAndWhere instead.
	SetWhere(from string) *gorm.DB
	SetUserAndWhere(user, where string) *gorm.DB
	GetRecords(objectId string) ([]*ChangeLog, error)
}

type Option func(options *options)

type plugin struct {
	db   *gorm.DB
	mu   sync.Mutex
	opts options
}

func Register(db *gorm.DB, opts ...Option) (LoggablePlugin, error) {
	err := db.AutoMigrate(&ChangeLog{}).Error
	if err != nil {
		return nil, err
	}
	o := options{}
	for _, option := range opts {
		option(&o)
	}
	r := &plugin{db: db, opts: o}
	callback := db.Callback()
	callback.Create().After("gorm:after_create").Register("loggable:create", r.addCreated)
	callback.Update().After("gorm:after_update").Register("loggable:update", r.addUpdated)
	callback.Delete().After("gorm:after_delete").Register("loggable:delete", r.addDeleted)
	return r, nil
}

func (r *plugin) GetRecords(objectId string) ([]*ChangeLog, error) {
	var changes []*ChangeLog
	return changes, r.db.Where("object_id = ?", objectId).Find(&changes).Error
}

func (r *plugin) getLastRecord(objectId string) (*ChangeLog, error) {
	var change ChangeLog
	return &change, r.db.Where("object_id = ?", objectId).Order("created_at DESC").Limit(1).Find(&change).Error
}

// Deprecated: Use SetUserAndWhere instead.
func (r *plugin) SetUser(user string) *gorm.DB {
	r.mu.Lock()
	db := r.db.Set("loggable:user", user)
	r.mu.Unlock()
	return db
}

// Deprecated: Use SetUserAndWhere instead.
func (r *plugin) SetWhere(where string) *gorm.DB {
	r.mu.Lock()
	db := r.db.Set("loggable:where", where)
	r.mu.Unlock()
	return db
}

func (r *plugin) SetUserAndWhere(user, where string) *gorm.DB {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.db.Set("loggable:user", user).Set("loggable:where", where)
}

func (r *plugin) addRecord(scope *gorm.Scope, action string) error {
	var jsonObject JSONB
	j, err := json.Marshal(scope.Value)
	if err != nil {
		return err
	}
	err = jsonObject.Scan(j)
	if err != nil {
		return err
	}
	user, where := getUserAndWhere(scope)

	cl := ChangeLog{
		ID:           uuid.NewV4().String(),
		ChangedBy:    user.(string),
		ChangedWhere: where.(string),
		Action:       action,
		ObjectID:     scope.PrimaryKeyValue().(string),
		ObjectType:   scope.GetModelStruct().ModelType.Name(),
		Object:       jsonObject,
	}
	return scope.DB().Create(&cl).Error
}

func getUserAndWhere(scope *gorm.Scope) (interface{}, interface{}) {
	user, ok := scope.DB().Get("loggable:user")
	if !ok {
		user = ""
	}
	where, ok := scope.DB().Get("loggable:where")
	if !ok {
		where = ""
	}
	return user, where
}

func isLoggable(scope *gorm.Scope) (isLoggable bool) {
	if scope.GetModelStruct().ModelType == nil {
		return false
	}
	_, isLoggable = reflect.New(scope.GetModelStruct().ModelType).Interface().(loggableInterface)
	return
}

func isEnabled(scope *gorm.Scope) (isEnabled bool) {
    if !isLoggable(scope) {
        return false
    }
    return scope.Value.(loggableInterface).Enabled()
}

func (r *plugin) addCreated(scope *gorm.Scope) {
	if isLoggable(scope) && isEnabled(scope) {
		r.addRecord(scope, "create")
	}
}

func (r *plugin) addUpdated(scope *gorm.Scope) {
	if isLoggable(scope) && isEnabled(scope) {
		if r.opts.lazyUpdate {
			record, err := r.getLastRecord(scope.PrimaryKeyValue().(string))
			if err == nil {
				if isEqual(record.Object, scope.Value, r.opts.lazyUpdateFields...) {
					return
				}
			}
		}
		r.addRecord(scope, "update")
	}
}

func (r *plugin) addDeleted(scope *gorm.Scope) {
	if isLoggable(scope) && isEnabled(scope) {
		r.addRecord(scope, "delete")
	}
}
