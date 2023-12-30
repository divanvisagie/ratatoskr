package layers

import (
	"ratatoskr/pkg/types"
	"ratatoskr/pkg/utils"
)

type CapabilitySelector struct {
	caps []types.Capability
	logger *utils.Logger
}

func NewCapabilitySelector(caps []types.Capability) *CapabilitySelector {
	logger := utils.NewLogger("Capability Selector")
	return &CapabilitySelector{caps, logger}
}

func (c *CapabilitySelector) PassThrough(req *types.RequestMessage) (types.ResponseMessage, error) {
	c.logger.Info("Capability selector layer: %+v\n", req.Message)
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
	c.logger.Info("Best capability: %+v\n", bestCapability)
	return bestCapability.Execute(req)
}
