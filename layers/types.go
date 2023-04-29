package layers

import "ratatoskr/types"

type Layer interface {
	PassThrough(*types.RequestMessage) (types.ResponseMessage, error)
}
