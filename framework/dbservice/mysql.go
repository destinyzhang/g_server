package dbservice

import (
	"database/sql"
	"g_server/framework/log"
	_ "github.com/go-sql-driver/mysql"
)

type MysqlService struct {
	db      *sql.DB
	address string
}

func (service *MysqlService) initDb() error {
	if service.db != nil {
		return nil
	}
	c, err := sql.Open("mysql", service.address)
	if err == nil {
		service.db = c
		service.db.SetMaxIdleConns(DbMaxCon)
	} else {
		glog.LogConsole(glog.LogError, "MysqlService initDb ", err)
	}
	return err
}

func (service *MysqlService) cmdEnd() {

}

func (service *MysqlService) iClose() {
	service.db.Close()
	service.db = nil
}

func (service *MysqlService) connInfo() string {
	return service.address
}

func (service *MysqlService) GetType() int {
	return DbMysql
}

func (service *MysqlService) GetMysqlDb() *sql.DB {
	return service.db
}
