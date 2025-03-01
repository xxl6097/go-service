package gore

type GlobService interface {
	IService
	OnUpgrade(string, string) []string
	OnInstall(string, string) []string
}

func Upgrade(svr IService, oldBinPath string, newFileUrlOrLocalPath string) []string {
	if gs, ok := svr.(GlobService); ok {
		return gs.OnUpgrade(oldBinPath, newFileUrlOrLocalPath)
	}
	return manualUpgrade(oldBinPath, newFileUrlOrLocalPath)
}

func manualUpgrade(oldBinPath string, newFileUrlOrLocalPath string) []string {
	return nil
}

func Installs(svr IService, binPath, installBinPath string) []string {
	if gs, ok := svr.(GlobService); ok {
		return gs.OnInstall(binPath, installBinPath)
	}
	return manualInstall(binPath, installBinPath)
}

func manualInstall(binPath, installBinPath string) []string {
	return nil
}
