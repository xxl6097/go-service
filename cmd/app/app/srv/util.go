package srv

import (
	"github.com/xxl6097/glog/pkg/z"
	"github.com/xxl6097/go-service/pkg"
	"github.com/xxl6097/go-service/pkg/ukey"
	"go.uber.org/zap"
)

func load() (*Config, error) {
	byteArray, err := ukey.Load()
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = ukey.GobToStruct(byteArray, &cfg)
	//err = json.Unmarshal(byteArray, &cfg)
	if err != nil {
		z.L().Error("ClientConfig解析错误", zap.Error(err))
		return nil, err
	}
	pkg.Version()
	return &cfg, nil
}
