package dbservice

import (
	"g_server/framework/log"
	"github.com/garyburd/redigo/redis"
)

type RedisService struct {
	db      *redis.Pool
	address string
	curconn redis.Conn
}

func (service *RedisService) dial() (redis.Conn, error) {
	c, err := redis.Dial("tcp", service.address)
	if err != nil {
		glog.LogConsole(glog.LogError, "RedisService dial ", err)
	}
	return c, err
}

func (service *RedisService) initDb() error {
	if service.db != nil {
		return nil
	}
	service.db = redis.NewPool(service.dial, DbMaxCon)
	return nil
}

func (service *RedisService) setCurconn() {
	if service.curconn == nil {
		service.curconn = service.db.Get()
	}
}

func (service *RedisService) cmdEnd() {
	if service.curconn != nil {
		service.curconn.Close()
		service.curconn = nil
	}
}

func (service *RedisService) iClose() {
	service.db.Close()
	service.db = nil
}

func (service *RedisService) connInfo() string {
	return service.address
}

func (service *RedisService) GetType() int {
	return DbRedis
}

func (service *RedisService) GetRedisDb() redis.Conn {
	service.setCurconn()
	return service.curconn
}
