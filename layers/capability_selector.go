package layers

import (
	"ratatoskr/types"
)

type CapabilitySelector struct {
	capabilities []types.Capability
}

func NewCapabilitySelector(capabilities []types.Capability) *CapabilitySelector {
	return &CapabilitySelector{capabilities}
}

func (c *CapabilitySelector) PassThrough(req *types.RequestMessage) (types.ResponseMessage, error) {
	for _, capability := range c.capabilities {
		if capability.Check(req) {
			return capability.Execute(req)
		}
	}
	return types.ResponseMessage{}, nil
}
