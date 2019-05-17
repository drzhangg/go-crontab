package common

import "github.com/pkg/errors"

var(
	ERR_LOCK_ALERADY_REQUIRED = errors.New("锁已被占用")

	ERR_NO_LOCAL_IP_FOUND = errors.New("没有找到网卡IP")
)
