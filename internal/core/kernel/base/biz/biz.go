package biz

import (
	"os"
	"sync"
	"time"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"supos.ai/operator/database/internal/config"
	"supos.ai/operator/database/internal/core/base/biz"
	"supos.ai/operator/database/pkg/common"
)

type changeCollection struct {
	catalog2ServiceList common.Catalog2ServiceList
	configPtr           *config.CfgItem
	startTime           time.Time
	loopCount           int
}

type Base struct {
	biz.Base

	curConfigFileInfo os.FileInfo

	curConfigPtr *config.CfgItem

	curRedoCheckMutex         sync.RWMutex
	curRedoCollection         []*changeCollection
	curCheckStatusRunningFlag bool
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *Base {
	ptr := &Base{
		Base:              biz.New(common.BaseModule, eventHub, backgroundRoutine),
		curRedoCollection: []*changeCollection{},
	}

	ptr.SubscribeFunc(common.NotifyTimer, ptr.timerCheck)
	ptr.SubscribeFunc(common.NotifyService, ptr.serviceNotify)

	return ptr
}

func (s *Base) timerCheck(_ event.Event, _ event.Result) {
}

func (s *Base) serviceNotify(ev event.Event, _ event.Result) {
}
