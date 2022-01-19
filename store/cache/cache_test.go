package cache_test

import (
	"fmt"
	"github.com/greywords/utils/store/cache"
	"log"
	"testing"
)

func init() {
	cache.InitRedisPool(&cache.Redis{
		Host: "127.0.0.1:6379",
	})
}

func TestSet(t *testing.T) {
	r, err := cache.Incr("test")
	print(r, err)
}

func TestGetString(t *testing.T) {
	r, err := cache.GetString("test1")
	log.Println(r, err)
}

func TestZAdd(t *testing.T) {
	_, err := cache.ZAdd("richlist", "pys", 10000)
	if err != nil {
		t.Fatal(err)
	}
}

func TestZRangeWithScores(t *testing.T) {
	list, err := cache.ZRangeWithScores("richlist", 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range list {
		println(v.Key, v.Score)
	}
}
func TestZRevRangeWithScores(t *testing.T) {
	list, err := cache.ZRevRangeWithScores("richlist", 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range list {
		println(v.Key, v.Score)
	}
}

func TestZRange(t *testing.T) {
	list, err := cache.ZRange("richlist", 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range list {
		println(k, v)
	}
}

func TestPublish(t *testing.T) {
	err := cache.Publish("lobby", "haha")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSets(t *testing.T) {
	//cache.IsExist("test1111")
	n, err := cache.SAdd("test1111", "aaab")
	fmt.Println(n, err)
}
