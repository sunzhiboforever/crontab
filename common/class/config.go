package class

import (
	"encoding/json"
	"io/ioutil"
)

var InstanceConfig *Config

//@todo 同时支持字符串配置参数
//@todo 支持多个配置文件的类型
type Config struct {
	ApiPort int	`json:"api_port"`
	ApiWriteTimeout int	`json:"api_write_timeout"`
	ApiReadTimeout int	`json:"api_read_timeout"`
	EtcdEndPoint []string `json:"etcd_end_point"`
	EtcdDialTimeout int `json:"etcd_dial_timeout"`
}

// 加载配置文件
func InitConfig(filename string) error {
	// 把配置文件读进来
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// 序列化json
	err = json.Unmarshal(content, &InstanceConfig)
	if err != nil {
		return err
	}
	return nil
}