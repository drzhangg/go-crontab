package common

import "github.com/pkg/errors"

var(
	ERR_LOCK_ALERADY_REQUIRED = errors.New("锁已被占用")
)
