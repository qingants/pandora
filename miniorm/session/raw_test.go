package session

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/qingants/pandora/miniorm/dialect"
	"github.com/qingants/pandora/miniorm/log"
)

var TestDB *sql.DB

func TestMain(m *testing.M) {
	log.Info("---->TestMain")
	var err error
	TestDB, err = sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/metrics")
	if err != nil {
		log.Error("connect data error")
	}
	if err = TestDB.Ping(); err != nil {
		log.Error(err)
		return
	}
	code := m.Run()
	_ = TestDB.Close()
	os.Exit(code)
}

func NewSession() *Session {
	dialect, _ := dialect.GetDialect("mysql")
	return New(TestDB, dialect)
}

func TestSession_Exec(t *testing.T) {
	s := NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS `ttask`").Exec()
	_, _ = s.Raw("CREATE TABLE IF NOT EXISTS `ttask` (`id` INT PRIMARY KEY NOT NULL AUTO_INCREMENT, `name` text);").Exec()
	result, _ := s.Raw("INSERT INTO `ttask` (`name`) values (?), (?)", "Go", "Python").Exec()
	if count, err := result.RowsAffected(); err != nil || count != 2 {
		t.Fatal("expect 2, but got", count)
	}
}

func TestSession_QueryRow(t *testing.T) {
	s := NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS `ttask`").Exec()
	_, _ = s.Raw("CREATE TABLE IF NOT EXISTS `ttask` (`id` INT PRIMARY KEY NOT NULL AUTO_INCREMENT, `name` text);").Exec()
	result, _ := s.Raw("INSERT INTO `ttask` (`name`) values (?), (?)", "Go", "Python").Exec()
	if count, err := result.RowsAffected(); err != nil || count != 2 {
		t.Fatal("expect 2, but got", count)
	}

	row := s.Raw("SELECT count(*) FROM `ttask`").QueryRow()
	var count int
	if err := row.Scan(&count); err != nil || count != 2 {
		t.Fatal("failed to query db", err)
	}
}

func TestSession_QueryRows(t *testing.T) {

}
