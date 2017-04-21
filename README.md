# Loggable

Loggable is used to helps tracking changes and history of your [GORM](https://github.com/jinzhu/gorm) models.

It creates `change_logs` table in your database and writes to all loggable models when changed.


## Usage

Just add `loggable.LoggableModel` to your GORM model.

If you want to set user, who make changes, use `loggable.SetUser`.
