package g

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"

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
