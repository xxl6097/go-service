package gore

type InstallService interface {
	IService
	OnInstall(string, string) []string
	GetBuffer() []byte
}

func Installs(svr IService, binPath, installBinPath string) []string {
	if gs, ok := svr.(InstallService); ok {
		buffer := gs.GetBuffer()
		if buffer != nil {
			return gs.OnInstall(binPath, installBinPath)
		}
	}
	return manualInstall(binPath, installBinPath)
}

func manualInstall(binPath, installBinPath string) []string {
	return nil
}

//安装、升级、卸载、开启、关闭、重启分别定义接口并继承BaseService
//1、安装，无实现，执行默认行为1（复制）；有实现，且参数合法，则执行子类行为，否则执行默认行为1
//2、作为go-service，默认行为2需要做签名功能
