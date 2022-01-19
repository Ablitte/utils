package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// type Error string

// func (err Error) Error() string { return string(err) }

// 字符串转为url.URL
func ParseEndpoint(endpoint string) (*url.URL, error) {
	endpoint = strings.Trim(endpoint, " ")
	endpoint = strings.TrimRight(endpoint, "/")
	if len(endpoint) == 0 {
		return nil, fmt.Errorf("empty URL")
	}
	i := strings.Index(endpoint, "://")
	if i >= 0 {
		scheme := endpoint[:i]
		if scheme != "http" && scheme != "https" {
			return nil, fmt.Errorf("invalid scheme: %s", scheme)
		}
	} else {
		endpoint = "http://" + endpoint
	}

	return url.ParseRequestURI(endpoint)
}

// 生成随机字符串
// length 32
//func GenerateRandomLenString() (string, error) {
//	length := 32
//	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
//	l := len(chars)
//	result := make([]byte, length)
//	_, err := rand.Read(result)
//	if err != nil {
//		return "", fmt.Errorf("Error reading random bytes: %v", err)
//	}
//	for i := 0; i < length; i++ {
//		result[i] = chars[int(result[i])%l]
//	}
//	return string(result), nil
//}

func GenerateRandomString(length uint8) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	l := len(chars)
	result := make([]byte, length)
	_, err := rand.Read(result)
	if err != nil {
		return "", fmt.Errorf("Error reading random bytes: %v", err)
	}
	for i := 0; i < int(length); i++ {
		result[i] = chars[int(result[i])%l]
	}
	return string(result), nil
}

func GenerateRandomNumber(length uint8) (string, error) {
	const chars = "0123456789"
	l := len(chars)
	result := make([]byte, length)
	_, err := rand.Read(result)
	if err != nil {
		return "", fmt.Errorf("Error reading random bytes: %v", err)
	}
	for i := 0; i < int(length); i++ {
		result[i] = chars[int(result[i])%l]
	}
	return string(result), nil
}

// 测试TCP连接
// timeout: the total time before returning if something is wrong
// with the connection, in second
// interval: the interval time for retring after failure, in second
func TestTCPConn(addr string, timeout, interval int) error {
	success := make(chan int)
	cancel := make(chan int)

	go func() {
		for {
			select {
			case <-cancel:
				break
			default:
				conn, err := net.DialTimeout("tcp", addr, time.Duration(timeout)*time.Second)
				if err != nil {
					//fmt.Errorf("failed to connect to tcp://%s, retry after %d seconds :%v",
					//	addr, interval, err)
					time.Sleep(time.Duration(interval) * time.Second)
					continue
				}
				//if err = conn.Close(); err != nil {
				//	log.Errorf("failed to close the connection: %v", err)
				//}
				conn.Close()
				success <- 1
				break
			}
		}
	}()

	select {
	case <-success:
		return nil
	case <-time.After(time.Duration(timeout) * time.Second):
		cancel <- 1
		return fmt.Errorf("failed to connect to tcp:%s after %d seconds", addr, timeout)
	}
}

// 时间戳秒数转时间
func ParseTimeStamp(timestamp string) (*time.Time, error) {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, err
	}
	t := time.Unix(i, 0)
	return &t, nil
}

// 通过json将map转为结构体
func ConvertMapToStruct(object interface{}, values interface{}) error {
	if object == nil {
		return errors.New("nil struct is not supported")
	}

	if reflect.TypeOf(object).Kind() != reflect.Ptr {
		return errors.New("object should be referred by pointer")
	}

	bytes, err := json.Marshal(values)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, object)
}

func ConvertStructToMap(object interface{}, value interface{}) error {
	if object == nil {
		return errors.New("nil struct is not supported")
	}

	if reflect.TypeOf(object).Kind() != reflect.Ptr {
		return errors.New("object should be referred by pointer")
	}

	bytes, err := json.Marshal(object)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, value)
}

// 判断文件路径是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//"application/x-www-form-urlencoded;charset=utf-8"
func DoPost(url, contentType, content string, timeout time.Duration) ([]byte, error) {
	client := &http.Client{
		Timeout: timeout,
	}
	if url == "" {
		return nil, errors.New("post: request url is empty")
	}
	if contentType == "" {
		contentType = "application/json;charset=utf-8"
	}
	return doPost(client, url, contentType, content)
}

func DoPostWithCert(url, contentType, content string, timeout time.Duration, cert tls.Certificate) ([]byte, error) {
	if url == "" {
		return nil, errors.New("post: request url is empty")
	}
	if contentType == "" {
		contentType = "application/json;charset=utf-8"
	}
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: true,
			},
		},
	}
	return doPost(client, url, contentType, content)
}
func doPost(client *http.Client, url, contentType, content string) ([]byte, error) {
	resp, err := client.Post(url,
		contentType,
		strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return rb, nil
}

func DoGet(url string, values url.Values, timeout time.Duration) ([]byte, error) {
	client := &http.Client{
		Timeout: timeout,
	}
	if url == "" {
		return nil, errors.New("get: request url is empty")
	}
	resp, err := client.Get(url + "?" + values.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rb, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}
	return rb, nil
}

func LoadFile(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func CallStack() string {
	buf := make([]byte, 10000)
	n := runtime.Stack(buf, false)
	buf = buf[:n]

	s := string(buf)

	const skip = 7
	count := 0
	index := strings.IndexFunc(s, func(c rune) bool {
		if c != '\n' {
			return false
		}
		count++
		return count == skip
	})
	return s[index+1:]
}

func GetHmacCode(s string) string {
	h := hmac.New(sha256.New, []byte("ourkey"))
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func GetSha256Code(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func GetMD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

//func GetUUID() string {
//	return strings.Replace(uuid.NewUUID(), "-", "", -1)
//}
