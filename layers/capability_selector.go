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
	bestScore := float32(0)
	var bestCapability types.Capability

	for _, capability := range c.caps {
		if score := capability.Check(req); score > bestScore {
			bestScore = score
			bestCapability = capability
			if bestScore == 1 {
				return capability.Execute(req)
			}
		}
	}
	return bestCapability.Execute(req)
}
