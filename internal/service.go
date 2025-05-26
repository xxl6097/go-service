package internal

import (
	"context"
	"github.com/xxl6097/go-service/pkg/utils"
)

func (this *CoreService) Upgrade(ctx context.Context, binUrl string) error {
	return this.upgrade(ctx, binUrl)
}

func (this *CoreService) UnInstall() error {
	//return this.uninstall()
	return utils.RunCmdBySelf("uninstall")
}

func (this *CoreService) RunCMD(args ...string) error {
	return utils.RunCmdBySelf(args...)
}

func (this *CoreService) Restart() error {
	return utils.RunCmdBySelf("restart")
}
