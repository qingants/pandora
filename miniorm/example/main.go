package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/qingants/pandora/miniorm"
	"github.com/qingants/pandora/miniorm/log"
)

func main() {
	engine, _ := miniorm.NewEngine("mysql", "root:root@tcp(127.0.0.1:3306)/metrics")
	defer engine.Close()

	session := engine.NewSession()
	_, _ = session.Raw("DROP TABLE IF EXISTS `user`;").Exec()
	_, _ = session.Raw("CREATE TABLE IF NOT EXISTS `user` (`ID` INT primary key NOT NULL AUTO_INCREMENT, `Name` text);").Exec()
	// _, _ = session.Raw("INSERT INTO `User` (`ID`, `Name`) VALUES (1, 'rocky')").Exec()
	result, _ := session.Raw("INSERT INTO `User` (`Name`) values (?), (?)", "Go", "Python").Exec()
	count, _ := result.RowsAffected()
	log.Info("Exec susccess, %d affected\n", count)
}
