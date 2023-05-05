package layers

import (
	"ratatoskr/types"
)

type CapabilitySelector struct {
	caps []types.Capability
}

func NewCapabilitySelector(caps []types.Capability) *CapabilitySelector {
	return &CapabilitySelector{caps}
}

func (c *CapabilitySelector) PassThrough(req *types.RequestMessage) (types.ResponseMessage, error) {
	for _, capability := range c.caps {
		if capability.Check(req) {
			return capability.Execute(req)
		}
	}
	return types.ResponseMessage{}, nil
}
