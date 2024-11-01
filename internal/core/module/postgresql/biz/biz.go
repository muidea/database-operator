package biz

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"supos.ai/operator/database/internal/core/base/biz"
	"supos.ai/operator/database/pkg/common"
)

type PostgreSQL struct {
	biz.Base
}

func New(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine,
) *PostgreSQL {
	ptr := &PostgreSQL{
		Base: biz.New(common.PostgreSQLModule, eventHub, backgroundRoutine),
	}

	return ptr
}
