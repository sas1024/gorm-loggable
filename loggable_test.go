package loggable

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB

type SomeType struct {
	gorm.Model
	Source string
	MetaModel
}

type MetaModel struct {
	createdBy string
	LoggableModel
}

func (m MetaModel) Meta() interface{} {
	return struct {
		CreatedBy string
	}{CreatedBy: m.createdBy}
}

func TestMain(m *testing.M) {
	database, err := gorm.Open(
		"postgres",
		fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			"root",
			"keepitsimple",
			"localhost",
			5432,
			"loggable",
		),
	)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	database = database.LogMode(true)
	_, err = Register(database)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	err = database.AutoMigrate(SomeType{}).Error
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	db = database
	os.Exit(m.Run())
}

func TestTryModel(t *testing.T) {
	newmodel := SomeType{Source: time.Now().Format(time.Stamp)}
	newmodel.createdBy = "some user"
	err := db.Create(&newmodel).Error
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(newmodel.ID)
	newmodel.Source = "updated field"
	err = db.Model(SomeType{}).Save(&newmodel).Error
	if err != nil {
		t.Fatal(err)
	}
}
