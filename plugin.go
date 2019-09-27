package loggable

import (
	"github.com/jinzhu/gorm"
)

// Plugin is a hook for gorm.
type Plugin struct {
	db   *gorm.DB
	opts options
}

// Register initializes Plugin for provided gorm.DB.
// There is also available some options, that should be passed there.
// Options cannot be set after initialization.
func Register(db *gorm.DB, opts ...Option) (Plugin, error) {
	err := db.AutoMigrate(&ChangeLog{}).Error
	if err != nil {
		return Plugin{}, err
	}
	o := options{}
	for _, option := range opts {
		option(&o)
	}
	p := Plugin{db: db, opts: o}
	callback := db.Callback()
	callback.Query().After("gorm:after_query").Register("loggable:query", p.trackEntity)
	callback.Create().After("gorm:after_create").Register("loggable:create", p.addCreated)
	callback.Update().After("gorm:after_update").Register("loggable:update", p.addUpdated)
	callback.Delete().After("gorm:after_delete").Register("loggable:delete", p.addDeleted)
	return p, nil
}

// GetRecords returns all records by objectId.
// Flag prepare allows to decode content of Raw* fields to direct fields, e.g. RawObject to Object.
func (p *Plugin) GetRecords(objectId string, prepare bool) (changes []ChangeLog, err error) {
	defer func() {
		if prepare {
			for i := range changes {
				if t, ok := p.opts.metaTypes[changes[i].ObjectType]; ok {
					err = changes[i].prepareMeta(t)
					if err != nil {
						return
					}
				}
				if t, ok := p.opts.objectTypes[changes[i].ObjectType]; ok {
					err = changes[i].prepareObject(t)
					if err != nil {
						return
					}
				}
			}
		}
	}()
	return changes, p.db.Where("object_id = ?", objectId).Find(&changes).Error
}

// GetLastRecord returns last by creation time (CreatedAt field) change log by provided object id.
// Flag prepare allows to decode content of Raw* fields to direct fields, e.g. RawObject to Object.
func (p *Plugin) GetLastRecord(objectId string, prepare bool) (change ChangeLog, err error) {
	defer func() {
		if prepare {
			if t, ok := p.opts.metaTypes[change.ObjectType]; ok {
				err := change.prepareMeta(t)
				if err != nil {
					return
				}
			}
			if t, ok := p.opts.objectTypes[change.ObjectType]; ok {
				err := change.prepareObject(t)
				if err != nil {
					return
				}
			}
		}
	}()
	return change, p.db.Where("object_id = ?", objectId).Order("created_at DESC").Limit(1).Find(&change).Error
}
