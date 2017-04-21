package loggable

import (
	"encoding/json"
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
)

func addRecord(db *gorm.DB, objectId string, object interface{}, action string) error {
	var jsonObject JSONB
	j, err := json.Marshal(object)
	if err != nil {
		return err
	}
	err = jsonObject.Scan(j)
	if err != nil {
		return err
	}
	cl := ChangeLog{
		ID:        uuid.NewV4().String(),
		ChangedBy: getUser(db),
		Action:    action,
		ObjectID:  objectId,
		Object:    jsonObject,
	}
	err = db.Create(&cl).Error
	if err != nil {
		return err
	}
	return nil
}

func GetRecords(db *gorm.DB, objectId string) ([]ChangeLog, error) {
	var changes []ChangeLog
	err := db.Find(&changes).Where("object_id = ?", objectId).Error
	if err != nil {
		return []ChangeLog{}, err
	}
	return changes, nil
}

func SetUser(db *gorm.DB, userId string) {
	db.InstantSet("loggable:userId", userId)
}

func getUser(db *gorm.DB) string {
	userId, ok := db.Get("loggable:userId")
	if !ok {
		return ""
	}
	return userId.(string)
}

func Register(db *gorm.DB) error {
	err := db.AutoMigrate(&ChangeLog{}).Error
	if err != nil {
		return err
	}
	callback := db.Callback()
	callback.Create().After("gorm:after_create").Register("changelog:create", addCreated)
	callback.Update().After("gorm:after_update").Register("changelog:update", addUpdated)
	callback.Delete().After("gorm:after_delete").Register("changelog:delete", addDeleted)
	return nil
}

func isLoggable(scope *gorm.Scope) (isLoggable bool) {
	if scope.GetModelStruct().ModelType == nil {
		return false
	}
	_, isLoggable = reflect.New(scope.GetModelStruct().ModelType).Interface().(loggableInterface)
	return
}

func addCreated(scope *gorm.Scope) {
	if isLoggable(scope) {
		addRecord(scope.DB(), scope.PrimaryKeyValue().(string), scope.Value, "create")
	}
}
func addUpdated(scope *gorm.Scope) {
	if isLoggable(scope) {
		addRecord(scope.DB(), scope.PrimaryKeyValue().(string), scope.Value, "update")
	}
}
func addDeleted(scope *gorm.Scope) {
	if isLoggable(scope) {
		addRecord(scope.DB(), scope.PrimaryKeyValue().(string), scope.Value, "delete")
	}
}
