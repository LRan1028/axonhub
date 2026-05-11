package objects

import (
	"slices"

	"github.com/shopspring/decimal"
)

type APIKeyProfiles struct {
	ActiveProfile string          `json:"activeProfile"`
	Profiles      []APIKeyProfile `json:"profiles"`
}

type APIKeyProfile struct {
	Name                string                         `json:"name"`
	ModelMappings       []ModelMapping                 `json:"modelMappings"`
	Quota               *APIKeyQuota                   `json:"quota,omitempty"`
	LoadBalanceStrategy *string                        `json:"loadBalanceStrategy,omitempty"`
	ModelAssociations   []APIKeyModelAssociationPolicy `json:"modelAssociations,omitempty"`
	StoragePolicy       *APIKeyStoragePolicy           `json:"storagePolicy,omitempty"`

	ChannelIDs           []int                `json:"channelIDs,omitempty"`
	ChannelTags          []string             `json:"channelTags,omitempty"`
	ChannelTagsMatchMode ChannelTagsMatchMode `json:"channelTagsMatchMode,omitempty"`
	ModelIDs             []string             `json:"modelIDs,omitempty"`
}

// ChannelTagsMatchMode controls how profile channel tags are matched.
// If this enum is changed, update MatchChannelTags in this file.
type ChannelTagsMatchMode string

const (
	ChannelTagsMatchModeAny  ChannelTagsMatchMode = "any"
	ChannelTagsMatchModeAll  ChannelTagsMatchMode = "all"
	ChannelTagsMatchModeNone ChannelTagsMatchMode = "none"
)

func (m ChannelTagsMatchMode) IsValid() bool {
	return m == "" || m == ChannelTagsMatchModeAny || m == ChannelTagsMatchModeAll || m == ChannelTagsMatchModeNone
}

func (m ChannelTagsMatchMode) OrDefault() ChannelTagsMatchMode {
	if m == ChannelTagsMatchModeAll {
		return ChannelTagsMatchModeAll
	}

	if m == ChannelTagsMatchModeNone {
		return ChannelTagsMatchModeNone
	}

	return ChannelTagsMatchModeAny
}

func (p *APIKeyProfile) MatchChannelTags(tags []string) bool {
	if p == nil || len(p.ChannelTags) == 0 {
		return true
	}

	return MatchChannelTags(p.ChannelTags, p.ChannelTagsMatchMode, tags)
}

func MatchChannelTags(allowedTags []string, matchMode ChannelTagsMatchMode, channelTags []string) bool {
	//nolint:exhaustive // Checked.
	switch matchMode.OrDefault() {
	case ChannelTagsMatchModeAll:
		for _, allowedTag := range allowedTags {
			if !slices.Contains(channelTags, allowedTag) {
				return false
			}
		}

		return true
	case ChannelTagsMatchModeNone:
		for _, tag := range channelTags {
			if slices.Contains(allowedTags, tag) {
				return false
			}
		}

		return true
	default:
		for _, tag := range channelTags {
			if slices.Contains(allowedTags, tag) {
				return true
			}
		}

		return false
	}
}

func (p *APIKeyProfile) Clone() *APIKeyProfile {
	if p == nil {
		return nil
	}
	cp := *p
	if len(p.ModelMappings) > 0 {
		cp.ModelMappings = make([]ModelMapping, len(p.ModelMappings))
		copy(cp.ModelMappings, p.ModelMappings)
	}
	if p.Quota != nil {
		q := *p.Quota
		if q.Requests != nil {
			r := *q.Requests
			q.Requests = &r
		}
		if q.TotalTokens != nil {
			tt := *q.TotalTokens
			q.TotalTokens = &tt
		}
		if q.Cost != nil {
			c := *q.Cost
			q.Cost = &c
		}
		q.Period = p.Quota.Period.clone()
		cp.Quota = &q
	}
	if len(p.ChannelIDs) > 0 {
		cp.ChannelIDs = make([]int, len(p.ChannelIDs))
		copy(cp.ChannelIDs, p.ChannelIDs)
	}
	if len(p.ChannelTags) > 0 {
		cp.ChannelTags = make([]string, len(p.ChannelTags))
		copy(cp.ChannelTags, p.ChannelTags)
	}
	if len(p.ModelIDs) > 0 {
		cp.ModelIDs = make([]string, len(p.ModelIDs))
		copy(cp.ModelIDs, p.ModelIDs)
	}
	if p.LoadBalanceStrategy != nil {
		s := *p.LoadBalanceStrategy
		cp.LoadBalanceStrategy = &s
	}
	if len(p.ModelAssociations) > 0 {
		cp.ModelAssociations = make([]APIKeyModelAssociationPolicy, len(p.ModelAssociations))
		for i := range p.ModelAssociations {
			cp.ModelAssociations[i] = p.ModelAssociations[i].clone()
		}
	}
	if p.StoragePolicy != nil {
		cp.StoragePolicy = p.StoragePolicy.clone()
	}
	return &cp
}

func (p *APIKeyQuotaPeriod) clone() APIKeyQuotaPeriod {
	if p == nil {
		return APIKeyQuotaPeriod{}
	}
	cp := *p
	if p.PastDuration != nil {
		pd := *p.PastDuration
		cp.PastDuration = &pd
	}
	if p.CalendarDuration != nil {
		cd := *p.CalendarDuration
		cp.CalendarDuration = &cd
	}
	return cp
}

type APIKeyQuota struct {
	Requests    *int64            `json:"requests,omitempty"`
	TotalTokens *int64            `json:"totalTokens,omitempty"`
	Cost        *decimal.Decimal  `json:"cost,omitempty"`
	Period      APIKeyQuotaPeriod `json:"period"`
}

type APIKeyModelAssociationPolicy struct {
	ModelID      string              `json:"modelId"`
	Associations []*ModelAssociation `json:"associations"`
}

func (p APIKeyModelAssociationPolicy) clone() APIKeyModelAssociationPolicy {
	cp := p
	if len(p.Associations) > 0 {
		cp.Associations = make([]*ModelAssociation, len(p.Associations))
		copy(cp.Associations, p.Associations)
	}
	return cp
}

func (p *APIKeyProfile) ModelAssociationPolicy(modelID string) *APIKeyModelAssociationPolicy {
	if p == nil || modelID == "" {
		return nil
	}

	for i := range p.ModelAssociations {
		if p.ModelAssociations[i].ModelID == modelID {
			return &p.ModelAssociations[i]
		}
	}

	return nil
}

type APIKeyStoragePolicy struct {
	DataStorageID     *int  `json:"dataStorageId,omitempty"`
	StoreChunks       *bool `json:"storeChunks,omitempty"`
	LivePreview       *bool `json:"livePreview,omitempty"`
	StoreRequestBody  *bool `json:"storeRequestBody,omitempty"`
	StoreResponseBody *bool `json:"storeResponseBody,omitempty"`
}

func (p *APIKeyStoragePolicy) clone() *APIKeyStoragePolicy {
	if p == nil {
		return nil
	}

	cp := *p
	if p.DataStorageID != nil {
		v := *p.DataStorageID
		cp.DataStorageID = &v
	}
	if p.StoreChunks != nil {
		v := *p.StoreChunks
		cp.StoreChunks = &v
	}
	if p.LivePreview != nil {
		v := *p.LivePreview
		cp.LivePreview = &v
	}
	if p.StoreRequestBody != nil {
		v := *p.StoreRequestBody
		cp.StoreRequestBody = &v
	}
	if p.StoreResponseBody != nil {
		v := *p.StoreResponseBody
		cp.StoreResponseBody = &v
	}

	return &cp
}

type APIKeyQuotaPeriod struct {
	Type             APIKeyQuotaPeriodType        `json:"type"`
	PastDuration     *APIKeyQuotaPastDuration     `json:"pastDuration,omitempty"`
	CalendarDuration *APIKeyQuotaCalendarDuration `json:"calendarDuration,omitempty"`
}

type APIKeyQuotaPeriodType string

const (
	APIKeyQuotaPeriodTypeAllTime          APIKeyQuotaPeriodType = "all_time"
	APIKeyQuotaPeriodTypePastDuration     APIKeyQuotaPeriodType = "past_duration"
	APIKeyQuotaPeriodTypeCalendarDuration APIKeyQuotaPeriodType = "calendar_duration"
)

type APIKeyQuotaPastDuration struct {
	Value int64                       `json:"value"`
	Unit  APIKeyQuotaPastDurationUnit `json:"unit"`
}

type APIKeyQuotaPastDurationUnit string

const (
	APIKeyQuotaPastDurationUnitMinute APIKeyQuotaPastDurationUnit = "minute"
	APIKeyQuotaPastDurationUnitHour   APIKeyQuotaPastDurationUnit = "hour"
	APIKeyQuotaPastDurationUnitDay    APIKeyQuotaPastDurationUnit = "day"
)

type APIKeyQuotaCalendarDuration struct {
	Unit APIKeyQuotaCalendarDurationUnit `json:"unit"`
}

type APIKeyQuotaCalendarDurationUnit string

const (
	APIKeyQuotaCalendarDurationUnitDay   APIKeyQuotaCalendarDurationUnit = "day"
	APIKeyQuotaCalendarDurationUnitMonth APIKeyQuotaCalendarDurationUnit = "month"
)
