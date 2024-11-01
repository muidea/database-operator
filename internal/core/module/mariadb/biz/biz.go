package biz

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"supos.ai/operator/database/internal/core/base/biz"
	"supos.ai/operator/database/pkg/common"
)

type Mariadb struct {
	biz.Base
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *Mariadb {
	ptr := &Mariadb{
		Base: biz.New(common.MariadbModule, eventHub, backgroundRoutine),
	}

	return ptr
}
