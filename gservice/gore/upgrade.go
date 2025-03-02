package gore

type GlobService interface {
	IService
	OnUpgrade(string, string) []string
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
