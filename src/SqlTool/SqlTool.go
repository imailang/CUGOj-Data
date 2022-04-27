package sqltool

import (
	properties "CUGOj-Data/src/Properties"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var sqlDB *sql.DB

func GetDB() *gorm.DB {
	return db
}
func InitialSql() error {
	user, err := properties.Get("SqlUsername")
	if err != nil {
		return err
	}
	password, err := properties.Get("SqlPassword")
	if err != nil {
		return err
	}
	ip, err := properties.Get("SqlIP")
	if err != nil {
		return err
	}
	port, err := properties.Get("SqlPort")
	if err != nil {
		return err
	}
	database, err := properties.Get("Database")
	if err != nil {
		return err
	}
	connectStr := user + ":" + password + "@tcp(" + ip + ":" + port + ")/" + database + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(connectStr), &gorm.Config{})
	if err != nil {
		return err
	}
	sqlDB, err = db.DB()
	if err != nil {
		return err
	}
	MaxIdleConns, err := properties.GetInt("MaxIdleConns")
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(MaxIdleConns)
	MaxOpenConns, err := properties.GetInt("MaxOpenConns")
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(MaxOpenConns)
	MaxLifeTime, err := properties.GetInt("MaxLifeTime")
	if err != nil {
		return err
	}
	sqlDB.SetConnMaxLifetime(time.Duration(MaxLifeTime))
	return nil
}

func CreateTables() error {
	err := db.AutoMigrate(
		&Problem{},
		&ProblemDescription{},
		&Judge{},
		&JudgeCase{},
		&Contest{},
		&ContestProblem{},
		&ContestRegister{},
		&ContestRecord{},
		&ContestJudge{},
		&ContestJudgeCase{},
		&UserInfo{},
	)
	if err == nil {
		fmt.Println("数据库表初始化成功")
	}
	return err
}

func QueryJudge(JudgeID string) (Judge, error) {
	judge := Judge{}
	res := db.Preload("UserInfo").Find(&judge, JudgeID)
	return judge, res.Error
}

func SaveJudge(judge *Judge) {
	db.Save(judge)
}

func QueryJudgeCases(JudgeID string) ([]JudgeCase, error) {
	judgeCases := []JudgeCase{}
	res := db.Where("j_id = ?", JudgeID).Find(&judgeCases)
	return judgeCases, res.Error
}
