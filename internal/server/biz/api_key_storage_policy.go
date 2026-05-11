package biz

import (
	"context"

	"github.com/looplj/axonhub/internal/contexts"
	"github.com/looplj/axonhub/internal/objects"
)

func EffectiveStoragePolicy(ctx context.Context, systemService *SystemService) *StoragePolicy {
	policy := defaultStoragePolicy
	if systemService != nil {
		policy = *systemService.StoragePolicyOrDefault(ctx)
	}

	override := activeAPIKeyStoragePolicy(ctx)
	if override == nil {
		return &policy
	}

	applyAPIKeyStoragePolicy(&policy, override)

	return &policy
}

func activeAPIKeyStoragePolicy(ctx context.Context) *objects.APIKeyStoragePolicy {
	apiKey, ok := contexts.GetAPIKey(ctx)
	if !ok || apiKey == nil {
		return nil
	}

	profile := apiKey.GetActiveProfile()
	if profile == nil {
		return nil
	}

	return profile.StoragePolicy
}

func applyAPIKeyStoragePolicy(policy *StoragePolicy, override *objects.APIKeyStoragePolicy) {
	if policy == nil || override == nil {
		return
	}

	if override.StoreChunks != nil {
		policy.StoreChunks = *override.StoreChunks
	}
	if override.LivePreview != nil {
		policy.LivePreview = *override.LivePreview
	}
	if override.StoreRequestBody != nil {
		policy.StoreRequestBody = *override.StoreRequestBody
	}
	if override.StoreResponseBody != nil {
		policy.StoreResponseBody = *override.StoreResponseBody
	}
}
