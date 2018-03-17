package g

import (
	"io/ioutil"
	"log"

	"time"

	yaml "gopkg.in/yaml.v2"
)

// Config global config struct
type Config struct {
	UUID   string //实例的UUID
	GameID string //游戏ID
	HAddr  string //HTTP rsync 监听地址
	HPort  string //HTTP rsync 监听端口
	QHost  string //redis 队列地址
	QPort  string //redis 队列端口
	QName  string //redis 队列名称
	XUser  string //MySQL 用户名
	XPass  string //MySQL 密码
	XPort  string //MySQL 端口
	XDirs  string //备份文件目录
}

//Xbackup backup config struct
type Xbackup struct {
	XUUID         string        //实例的UUID
	XAddr         string        //实例的IP地址
	XType         string        //实例类型(MySQL|Redis)
	XPort         string        //实例端口
	XDURL         string        //实例备份成功的下载地址
	XStat         bool          //实例是否备份成功
	XFilename     string        //实例备份成功后的文件名
	XLastFilename string        //实例上次备份成功的文件名
	XStartTime    time.Duration //实例备份的开始时间
	XDoneTime     time.Duration //实例备份的结束时间
}

//Plugin backup script
type Plugin struct {
	FilePath string
	Cycle    int
}

// PXConfig parse config file
func PXConfig(cfg string) *Config {
	var xConfig Config
	source, err := ioutil.ReadFile(cfg)
	if err != nil {
		log.Fatalln("")
	}
	err = yaml.Unmarshal(source, &xConfig)
	if err != nil {
		log.Fatalln("")
	}
	return &xConfig
}
