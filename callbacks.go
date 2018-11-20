package loggable

import (
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
)

func (p *Plugin) addCreated(scope *gorm.Scope) {
	if isLoggable(scope) && isEnabled(scope) {
		addRecord(scope, "create")
	}
}

func (p *Plugin) addUpdated(scope *gorm.Scope) {
	if isLoggable(scope) && isEnabled(scope) {
		if p.opts.lazyUpdate {
			record, err := p.GetLastRecord(interfaceToString(scope.PrimaryKeyValue()), false)
			if err == nil {
				if isEqual(record.RawObject, scope.Value, p.opts.lazyUpdateFields...) {
					return
				}
			}
		}
		addRecord(scope, "update")
	}
}

func (p *Plugin) addDeleted(scope *gorm.Scope) {
	if isLoggable(scope) && isEnabled(scope) {
		addRecord(scope, "delete")
	}
}

func addRecord(scope *gorm.Scope, action string) error {
	rawObject, err := json.Marshal(scope.Value)
	if err != nil {
		return err
	}
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	cl := ChangeLog{
		ID:         id,
		Action:     action,
		ObjectID:   interfaceToString(scope.PrimaryKeyValue()),
		ObjectType: scope.GetModelStruct().ModelType.Name(),
		RawObject:  string(rawObject),
		RawMeta:    string(fetchChangeLogMeta(scope)),
	}
	return scope.DB().Create(&cl).Error
}
