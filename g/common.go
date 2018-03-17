package g

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql" // user mysql driver
)

//IsExist insure dir exist
func isExist(dir string) bool {
	_, err := os.Stat(dir)
	return err == nil || os.IsExist(err)
}

//InitBackupDir insure backup dir exist or create one
func InitBackupDir(dir string) {
	if isExist(dir) {
		log.Printf("BackupDir exist: %s", dir)
	}
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Printf("Mkdir %s faild", dir)
	}
}

//InitMySQLAlive insure mysql alive
func InitMySQLAlive(user, pass, port, host string) {
	db, err := sql.Open("mysql", user+":"+pass+"@tcp("+host+":"+port+")")
	if err != nil {
		log.Fatalf("MySQL connect faild with user: %s, pass: %s, host: %s, port:%s", user, pass, host, port)
	}
	defer db.Close()
}

//InitQueueAlive insure redis queue alive
func InitQueueAlive(host, port string) {
	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port,
	})
	_, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("Redis connect faild with host: %s, port: %s", host, port)
	}
	defer client.Close()
}

//Md5Sum cal the file md5
func Md5Sum(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	checkSum := fmt.Sprintf("%x", hash.Sum(nil))
	return checkSum, nil
}

// GetLocalIPAddr get instance ipaddr
func GetLocalIPAddr() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return ""
}

//TimeToYmdHms format to 2006-01-02-15-04-05
func TimeToYmdHms() string {
	return time.Now().Format("2006-01-02-15-04-05")
}

// WriteBytes write file
func WriteBytes(filePath string, wb []byte) (int, error) {
	os.MkdirAll(path.Dir(filePath), os.ModePerm)
	fi, err := os.Create(filePath)
	if err != nil {
		return 0, err
	}
	defer fi.Close()
	return fi.Write(wb)
}

// WriteString write file
func WriteString(filePath, ws string) (int, error) {
	return WriteBytes(filePath, []byte(ws))
}

// ReadToByte read file
func ReadToByte(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

// ReadToString read file
func ReadToString(filePath string) (string, error) {
	bs, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
