package dbservice

import (
	"errors"
	"fmt"
	"g_server/framework/com"
	"g_server/framework/log"
)

const (
	DbMaxCon         = 5
	ServericeStop    = 0
	ServericeStoping = 1
	ServericeRun     = 2

	DbMysql = 1
	DbRedis = 2
)

var (
	ErrRedisConnNil = errors.New("Err RedisConnNil")
	ErrMysqlConnNil = errors.New("Err MysqlConnNil")
)

type DbServiceContiner struct {
	resultque   *com.SyncQueue
	state       int
	reuqestChan chan IDbRequest
	iService    IDbService
	name        string
}

func (service *DbServiceContiner) close() {
	service.reuqestChan = nil
	service.iService.iClose()
	service.state = ServericeStop
}

func (service *DbServiceContiner) run() {
	service.state = ServericeRun
	go func() {
		defer service.close()
		var (
			result IDbResult
			err    error
			retry  int
		)
		service.iService.initDb()
		for {
			reuqest, ok := <-service.reuqestChan
			if !ok {
				return
			}
			com.SafeCall(func() {
				retry = reuqest.RetryCount()
			Retry:
				if err, result = reuqest.DoCmd(service.iService); err != nil && retry > 0 {
					glog.LogConsole(glog.LogError, "reuqest Docmd err", err)
					retry--
					goto Retry
				}
				if result != nil {
					service.resultque.Push(result)
				}
			})
			service.iService.cmdEnd()
		}
	}()
}

func (service *DbServiceContiner) Run() {
	for {
		result := service.resultque.Pop()
		if result == nil {
			return
		}
		com.SafeCall(func() {
			result.(IDbResult).CallBack()
		})
	}
}

func (service *DbServiceContiner) Name() string {
	return service.name
}

func (service *DbServiceContiner) Start() bool {
	if service.state != ServericeStop {
		return false
	}
	service.reuqestChan = make(chan IDbRequest, 100)
	service.run()
	return true
}

func (service *DbServiceContiner) Stop() bool {
	if service.state == ServericeStop || service.state == ServericeStoping {
		return false
	}
	service.state = ServericeStoping
	close(service.reuqestChan)
	for {
		if service.state == ServericeStop {
			break
		}
	}
	return true
}

func (service *DbServiceContiner) PushRequest(request IDbRequest) {
	if service.state != ServericeRun {
		return
	}
	com.SafeGo(func() {
		service.reuqestChan <- request
	})
}

func (service *DbServiceContiner) CheckState(state int) bool {
	return service.state == state
}

func (service *DbServiceContiner) GetConninfo() string {
	return service.iService.connInfo()
}

func (service *DbServiceContiner) GetType() int {
	return service.iService.GetType()
}

type IDbRequest interface {
	DoCmd(IDbService) (error, IDbResult)
	RetryCount() int
}

type IDbResult interface {
	CallBack()
}

type IDbService interface {
	iClose()
	initDb() error
	cmdEnd()
	connInfo() string
	GetType() int
}

func NewRedisService(addr string, name string) *DbServiceContiner {
	return &DbServiceContiner{resultque: com.NewSyncQueue(), name: name, iService: &RedisService{address: addr}}
}

func NewMysqlService(dbusername string, dbpassword string, dbhostsip string, dbname string, name string) *DbServiceContiner {
	return &DbServiceContiner{resultque: com.NewSyncQueue(), name: name, iService: &MysqlService{address: fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", dbusername, dbpassword, dbhostsip, dbname)}}
}
