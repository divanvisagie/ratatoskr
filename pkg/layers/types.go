package layers

import "ratatoskr/pkg/types"

type Layer interface {
	PassThrough(*types.RequestMessage) (types.ResponseMessage, error)
}
