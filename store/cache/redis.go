package cache

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Redis
type Redis struct {
	Host string `json:"host"`
	DB   string `json:"db,omitempty"`
	Auth string `json:"auth,omitempty"`
}

// redis 连接池
var redisPool *redis.Pool
var prefixName string

func RedisPool() *redis.Pool {
	return redisPool
}

func SetPrefixName(name string) {
	prefixName = name
}

func GetPrefixName() string {
	return prefixName
}

// 创建redis 连接池
func InitRedisPool(rcnf *Redis) {
	dialFunc := func() (redis.Conn, error) {
		// Dial connection
		c, err := redis.Dial("tcp",
			rcnf.Host)
		if err != nil {
			return nil, err
		}
		// Auth
		if rcnf.Auth != "" {
			if _, err := c.Do("AUTH", rcnf.Auth); err != nil {
				c.Close()
				return nil, err
			}
		}
		// Select DB
		if rcnf.DB != "" {
			if _, err := c.Do("SELECT", rcnf.DB); err != nil {
				c.Close()
				return nil, err
			}
		}
		return c, nil
	}
	redisPool = &redis.Pool{
		Dial:            dialFunc,
		IdleTimeout:     time.Minute * 60,
		MaxConnLifetime: time.Minute * 5,
	}
}

// 执行redis命令
func Do(cmd string, args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("redigo do func missing required arguments")
	}
	c := redisPool.Get()
	defer c.Close()

	return c.Do(cmd, args...)
}

// 检测键是否存在
func IsExist(key string) bool {
	r, err := redis.Bool(Do("EXISTS", prefixName+key))
	if err != nil {
		return false
	}
	return r
}

// 永不过期
func SetNE(k string, v interface{}) error {
	_, err := Do("SET", prefixName+k, v)
	return err
}

//默认十分钟过期时间
func Set(k string, v interface{}) error {
	_, err := Do("SETEX", prefixName+k, int64(10*60), v)
	return err
}
func SetEX(k string, t time.Duration, v interface{}) error {
	_, err := Do("SETEX", prefixName+k, int64(t/time.Second), v)
	return err
}

func SetNX(k string, v interface{}) error {
	ret, err := redis.Int(Do("SETNX", prefixName+k, v))
	if err != nil {
		return err
	}
	if ret == 0 {
		kt := k
		if len(kt) > 15 {
			kt = kt[0:12] + "..."
		}
		return fmt.Errorf("error setnx key %s", kt)
	}
	return err
}

func LPush(k string, v ...interface{}) (int64, error) {
	var param []interface{}
	param = append(param, prefixName+k)
	param = append(param, v...)
	return redis.Int64(Do("LPUSH", param...))
}

//****************************SET****************************
func SAdd(k string, v ...interface{}) (int64, error) {
	var param []interface{}
	param = append(param, prefixName+k)
	param = append(param, v...)
	return redis.Int64(Do("SADD", param...))
}

func SRem(k string, v ...interface{}) (int64, error) {
	var param []interface{}
	param = append(param, prefixName+k)
	param = append(param, v...)
	return redis.Int64(Do("SREM", param...))
}

//查看set元素数量
func SCard(k string) (int64, error) {
	return redis.Int64(Do("SCARD", prefixName+k))
}

func GetBytes(k string) ([]byte, error) {
	r, err := redis.Bytes(Do("GET", prefixName+k))
	return r, err
}
func GetInt64(k string) (int64, error) {
	r, err := redis.Int64(Do("GET", prefixName+k))
	return r, err
}
func GetInt(k string) (int, error) {
	r, err := redis.Int(Do("GET", prefixName+k))
	return r, err
}
func GetString(k string) (string, error) {
	r, err := redis.String(Do("GET", prefixName+k))
	return r, err
}
func GetBool(k string) (bool, error) {
	r, err := redis.Bool(Do("GET", prefixName+k))
	return r, err
}

// 删除一个键
func Del(k string) error {
	_, err := Do("DEL", prefixName+k)
	return err
}

// 设置过期时间
func Expire(k string, t time.Duration) error {
	_, err := Do("EXPIRE", prefixName+k, int64(t/time.Second))
	return err
}

// 发布
func Publish(channel string, msg interface{}) error {
	_, err := Do("PUBLISH", channel, msg)
	return err
}

//增减
func Incr(k string) (int64, error) {
	return redis.Int64(Do("INCR", k))
}

func Decr(k string) (int64, error) {
	return redis.Int64(Do("DECR", k))
}

type ZRankItem struct {
	Key   string
	Score int64
}

//插入或更新元素
func ZAdd(rank string, k string, s int64) (int64, error) {
	return redis.Int64(Do("ZADD", prefixName+rank, s, k))
}

func ZIncrBy(rank string, k string, s int64) (int64, error) {
	return redis.Int64(Do("ZINCRBY", prefixName+rank, s, k))
}
func ZRank(rank string, k string) (int, error) {
	return redis.Int(Do("ZRANK", prefixName+rank, k))
}

func ZSCORE(rank string, k string) (int64, error) {
	return redis.Int64(Do("ZSCORE", prefixName+rank, k))
}

func ZRevRank(rank string, k string) (int, error) {
	return redis.Int(Do("ZREVRANK", prefixName+rank, k))
}

//删除元素
func ZRem(rank string, k string) error {
	_, err := Do("ZREM", prefixName+rank, k)
	return err
}

//TODO: 转map会有乱序的问题，暂时先用转list的方式代替
func ZRangeWithScores(rank string, start, stop int) ([]ZRankItem, error) {
	var itemList []ZRankItem
	ret, err := redis.Strings(Do("ZRANGE", prefixName+rank, start, stop, "WITHSCORES"))
	if err != nil {
		return itemList, err
	}
	for i := 0; i+1 < len(ret); i = i + 2 {
		score, err := strconv.ParseInt(ret[i+1], 10, 64)
		if err != nil {
			continue
		}
		itemList = append(itemList, ZRankItem{
			Key:   ret[i],
			Score: score,
		})
	}
	return itemList, nil
}

func ZRange(rank string, start, stop int) ([]string, error) {
	return redis.Strings(Do("ZRANGE", prefixName+rank, start, stop))
}

func ZRevRangeWithScores(rank string, start, stop int) ([]ZRankItem, error) {
	var itemList []ZRankItem
	ret, err := redis.Strings(Do("ZREVRANGE", prefixName+rank, start, stop, "WITHSCORES"))
	if err != nil {
		return itemList, err
	}
	for i := 0; i+1 < len(ret); i = i + 2 {
		score, err := strconv.ParseInt(ret[i+1], 10, 64)
		if err != nil {
			continue
		}
		itemList = append(itemList, ZRankItem{
			Key:   ret[i],
			Score: score,
		})
	}
	return itemList, nil
}

func ZRevRange(rank string, start, stop int) ([]string, error) {
	return redis.Strings(Do("ZREVRANGE", prefixName+rank, start, stop))
}

func ZRemRangeByScore(rank string, start, stop int64) (int, error) {
	return redis.Int(Do("ZREMRANGEBYSCORE", prefixName+rank, start, stop))
}

func ZRemRangeByRank(rank string, start, stop int64) (int, error) {
	return redis.Int(Do("ZREMRANGEBYRANK", prefixName+rank, start, stop))
}

//****************************HASH****************************

func HSetInt(rank, field string, value int) error {
	_, err := Do("HSET", prefixName+rank, field, value)
	return err
}

func HGetInt(rank, field string) (int, error) {
	return redis.Int(Do("HGET", prefixName+rank, field))
}

func HGetAllInt(rank string) (map[string]int, error) {
	return redis.IntMap(Do("HGETALL", prefixName+rank))
}

func HExistsField(rank, field string) bool {
	exists, err := redis.Bool(Do("HEXISTS", prefixName+rank, field))
	if err != nil {
		return false
	}
	return exists
}

func HDel(rank, field string) error {
	_, err := Do("HDEL", prefixName+rank, field)
	return err
}

func HGet(rank, field string, v interface{}) error {
	temp, _ := redis.Bytes(Do("HGET", prefixName+rank, field))

	return json.Unmarshal(temp, &v) // 反序列化
}

func HSet(rank string, field string, v interface{}) (err error) {
	value, err := json.Marshal(v)
	_, err = Do("HSET", prefixName+rank, field, value)

	return err
}

//****************************no prefix****************************
// get int no prefix
func GetIntNP(k string) (int, error) {
	r, err := redis.Int(Do("GET", k))
	return r, err
}

// set with ex, no prefix
func SetEXNP(k string, t time.Duration, v interface{}) error {
	_, err := Do("SETEX", k, int64(t/time.Second), v)
	return err
}

// del一个键
func DelNP(k string) error {
	_, err := Do("DEL", k)
	return err
}
