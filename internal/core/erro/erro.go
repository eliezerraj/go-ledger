package erro

import (
	"errors"
)

var (
	ErrNotFound 		= errors.New("item not found")
	ErrTypeInvalid 		= errors.New("moviment type invalid")
	ErrHTTPForbiden		= errors.New("forbiden request")
	ErrUnauthorized 	= errors.New("not authorized")
	ErrServer		 	= errors.New("server identified error")
	ErrUnmarshal 		= errors.New("unmarshal json error")
	ErrForceRollback 	= errors.New("force rollback")
	ErrTimeout			= errors.New("timeout: context deadline exceeded.")
)