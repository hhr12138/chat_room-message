package object

import (
	"encoding/json"
	"github.com/go-redis/redis"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/hhr12138/door/entity"
	"github.com/hhr12138/door/vo"
	"log"
	"time"
)

const (
	ARC_SIZE       = 10000
	CACHE_TIME_OUT = 7 * 24 * 60 * 60 * 1000 //ms
	REDIS_ADDR     = "127.0.0.1:6379"
	REDIS_PASSWD   = ""
	REDIS_DB       = 0
)

const(
	MAX_INF = "+inf"
	MIN_INF = "-inf"
)

var UserCache *lru.ARCCache[string, *vo.UserCache]
var RedisClient *redis.Client

func RegisterCache() {
	var err error
	UserCache, err = lru.NewARC[string, *vo.UserCache](ARC_SIZE)
	if err != nil {
		panic("arc register err: " + err.Error())
	}

	//初始化redis
	initRedisClient()
}

func initRedisClient() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     REDIS_ADDR,
		Password: REDIS_PASSWD,
		DB:       REDIS_DB,
	})
	_, err := RedisClient.Ping().Result()
	if err != nil {
		panic("redis client init err:" + err.Error())
	}
}

func CacheUser(token string, user *entity.User) {
	cacheUser := &vo.UserCache{
		User:         user,
		CacheTimeOut: time.Now().UnixMilli() + CACHE_TIME_OUT,
	}
	UserCache.Add(token, cacheUser)
	userbs, _ := json.Marshal(cacheUser)
	_, err := RedisClient.Set(token, string(userbs), CACHE_TIME_OUT*time.Millisecond).Result()
	if err != nil {
		//todo
		log.Fatalln(err)
	}
}

func GetUser(token string) (*entity.User, bool) {
	user, ok := getUserFromLocalCache(token)
	if !ok {
		user = getUserFromRedis(token)
	}
	if user != nil && user.User != nil && user.CacheTimeOut > time.Now().UnixMilli() {
		return user.User, true
	}
	return nil, false
}

func RemoveUser(token string) {
	UserCache.Remove(token)
	RedisClient.Del(token)
}

func getUserFromRedis(token string) *vo.UserCache {
	user := &vo.UserCache{}
	result, err := RedisClient.Get(token).Bytes()
	if err == nil {
		err = json.Unmarshal(result, user)
		if err != nil {
			log.Fatalln(err)
		}
	} else if err != redis.Nil {
		//todo: 改成给kafka发消息, 然后让java的日志处理模块处理
		log.Fatalln(err)
	}
	return user
}

func getUserFromLocalCache(token string) (*vo.UserCache, bool) {
	var ok bool
	var user *vo.UserCache
	user, ok = UserCache.Get(token)
	return user, ok
}
