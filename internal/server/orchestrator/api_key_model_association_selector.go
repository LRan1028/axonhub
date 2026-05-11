package orchestrator

import (
	"context"

	"github.com/samber/lo"

	"github.com/looplj/axonhub/internal/log"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/server/biz"
	"github.com/looplj/axonhub/llm"
)

// APIKeyModelAssociationSelector replaces the global model association lookup
// when the active API key profile defines a policy for the requested model.
type APIKeyModelAssociationSelector struct {
	fallback       CandidateSelector
	channelService *biz.ChannelService
	associations   []*objects.ModelAssociation
}

func NewAPIKeyModelAssociationSelector(
	fallback CandidateSelector,
	channelService *biz.ChannelService,
	associations []*objects.ModelAssociation,
) *APIKeyModelAssociationSelector {
	return &APIKeyModelAssociationSelector{
		fallback:       fallback,
		channelService: channelService,
		associations:   associations,
	}
}

func (s *APIKeyModelAssociationSelector) Select(ctx context.Context, req *llm.Request) ([]*ChannelModelsCandidate, error) {
	if len(s.associations) == 0 || s.channelService == nil {
		return s.fallback.Select(ctx, req)
	}

	channels := s.channelService.GetEnabledChannels()
	matches := biz.MatchAssociations(s.associations, channels)
	if len(matches) == 0 {
		return []*ChannelModelsCandidate{}, nil
	}

	channelMap := make(map[int]*biz.Channel, len(channels))
	for _, ch := range channels {
		channelMap[ch.ID] = ch
	}

	resolved := make([]*resolvedAssociationCandidate, 0, len(matches))
	for _, match := range matches {
		for _, conn := range match.Connections {
			ch := channelMap[conn.Channel.ID]
			if ch == nil {
				continue
			}

			resolved = append(resolved, &resolvedAssociationCandidate{
				channel:  ch,
				priority: conn.Priority,
				models:   append([]biz.ChannelModelEntry(nil), conn.Models...),
				when:     match.Association.When,
			})
		}
	}

	candidates := filterResolvedCandidatesForRequest(ctx, req, resolved)

	if log.DebugEnabled(ctx) {
		log.Debug(ctx, "selected api key model association candidates",
			log.String("model", req.Model),
			log.Int("association_count", len(s.associations)),
			log.Int("candidate_count", len(candidates)),
			log.Any("candidate_channel_ids", lo.Map(candidates, func(c *ChannelModelsCandidate, _ int) int {
				return c.Channel.ID
			})),
		)
	}

	return candidates, nil
}
