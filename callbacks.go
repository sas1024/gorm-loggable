package loggable

import (
	"encoding/json"
	"reflect"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
)

type LoggablePlugin interface {
	SetUser(user string) *gorm.DB
	GetRecords(objectId string) ([]*ChangeLog, error)
}

type loggablePlugin struct {
	db *gorm.DB
	mu sync.Mutex
}

func Register(db *gorm.DB) (LoggablePlugin, error) {
	err := db.AutoMigrate(&ChangeLog{}).Error
	if err != nil {
		return nil, err
	}
	r := &loggablePlugin{db: db}
	callback := db.Callback()
	callback.Create().After("gorm:after_create").Register("loggable:create", r.addCreated)
	callback.Update().After("gorm:after_update").Register("loggable:update", r.addUpdated)
	callback.Delete().After("gorm:after_delete").Register("loggable:delete", r.addDeleted)
	return r, nil
}

func (r *loggablePlugin) GetRecords(objectId string) ([]*ChangeLog, error) {
	var changes []*ChangeLog
	err := r.db.Find(&changes).Where("object_id = ?", objectId).Error
	if err != nil {
		return nil, err
	}
	return changes, nil
}

func (r *loggablePlugin) SetUser(user string) *gorm.DB {
	r.mu.Lock()
	db := r.db.Set("loggable:user", user)
	r.mu.Unlock()
	return db
}

func (r *loggablePlugin) addRecord(scope *gorm.Scope, action string) error {
	var jsonObject JSONB
	j, err := json.Marshal(scope.Value)
	if err != nil {
		return err
	}
	err = jsonObject.Scan(j)
	if err != nil {
		return err
	}
	user, ok := scope.DB().Get("loggable:user")
	if !ok {
		user = ""
	}

	cl := ChangeLog{
		ID:         uuid.NewV4().String(),
		ChangedBy:  user.(string),
		Action:     action,
		ObjectID:   scope.PrimaryKeyValue().(string),
		ObjectType: scope.GetModelStruct().ModelType.Name(),
		Object:     jsonObject,
	}
	err = scope.DB().Create(&cl).Error
	if err != nil {
		return err
	}
	return nil
}

func isLoggable(scope *gorm.Scope) (isLoggable bool) {
	if scope.GetModelStruct().ModelType == nil {
		return false
	}
	_, isLoggable = reflect.New(scope.GetModelStruct().ModelType).Interface().(loggableInterface)
	return
}

func (r *loggablePlugin) addCreated(scope *gorm.Scope) {
	if isLoggable(scope) {
		r.addRecord(scope, "create")
	}
}
func (r *loggablePlugin) addUpdated(scope *gorm.Scope) {
	if isLoggable(scope) {
		r.addRecord(scope, "update")
	}
}
func (r *loggablePlugin) addDeleted(scope *gorm.Scope) {
	if isLoggable(scope) {
		r.addRecord(scope, "delete")
	}
}
