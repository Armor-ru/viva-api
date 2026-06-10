package viva_api

import (
	"git.dev.armlab.pro/armor/sds-go/pkg/errs"
	"git.dev.armlab.pro/armor/sds-go/pkg/logger"
)

func logAppError(err error, msg string) {
	if err == nil {
		return
	}
	logger.Error().Fields(errs.Fields(err)).Err(err).Msg(msg)
}
