# Loggable

Loggable is used to helps tracking changes and history of your [GORM](https://github.com/jinzhu/gorm) models.

It creates `change_logs` table in your database and writes to all loggable models when they are changed.

More documentation is available in [godoc](https://godoc.org/github.com/sas1024/gorm-loggable).

## Usage
1. Register plugin using `loggable.Register(db)`.
```go
plugin, err := Register(database) // database is a *gorm.DB
if err != nil {
	panic(err)
}
```
2. Add (embed) `loggable.LoggableModel` to your GORM model.
```go
type User struct{
    Id        uint
    CreatedAt time.Time
    // some other stuff...
    
    loggable.LoggableModel
}
```
3. Changes after calling Create, Save, Update, Delete will be tracked.

## Customization
You may add additional fields to change logs, that should be saved.  
First, embed `loggable.LoggableModel` to your model wrapper or directly to GORM model.  
```go
type CreatedByLog struct {
	// Public field will be catches by GORM and will be saved to main table.
	CreatedBy     string
	// Hided field because we do not want to write this to main table,
	// only to change_logs.
	createdByPass string 
	loggable.LoggableModel
}
```
After that, shadow `LoggableModel`'s `Meta()` method by writing your realization, that should return structure with your information.  
```go
type CreatedByLog struct {
	CreatedBy     string
	createdByPass string 
	loggable.LoggableModel
}

func (m CreatedByLog) Meta() interface{} {
	return struct { // You may define special type for this purposes, here we use unnamed one.
		CreatedBy     string
		CreatedByPass string // CreatedByPass is a public because we want to track this field. 
	}{
		CreatedBy:     m.CreatedBy,
		CreatedByPass: m.createdByPass,
	}
}
```

## Options
#### LazyUpdate
Option `LazyUpdate` allows save changes only if they big enough to be saved.  
Plugin compares the last saved object and the new one, but ignores changes was made in fields from provided list.

#### ComputeDiff
Option `ComputeDiff` allows to only save the changes into the RawDiff field. This options is only relevant during update
operations. Only fields tagged with `gorm-loggable:true` will be taken in account. If the object does not have any field
tagged with `gorm-loggable:true` then the column will always be `NULL`.

e.g.

```go
type Person struct {
	FirstName string `gorm-loggable:true`
	LastName  string `gorm-loggable:true`
	Age       int    `gorm-loggable:true`
}
```

Let's say you change person `FirstName` from `John` to `Jack` and its `Age` from 30 to 40.
`ChangeLog.RawDiff` will be populated with the following:
```json
{
  "FirstName": "Jack",
  "Age": 40,
}
```
