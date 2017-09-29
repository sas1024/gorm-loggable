# Loggable

Loggable is used to helps tracking changes and history of your [GORM](https://github.com/jinzhu/gorm) models.

It creates `change_logs` table in your database and writes to all loggable models when they are changed.


## Usage
 1. Register plugin using `loggable.Register(db)`
 2. Add `loggable.LoggableModel` to your GORM model
 3. If you want to set user, who makes changes, and place, where it happened, use `loggable.SetUserAndFrom("username", "London")`.
