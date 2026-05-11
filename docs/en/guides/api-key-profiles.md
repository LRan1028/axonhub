# API Key Profile Guide

This guide explains how to configure API Key Profiles for model mapping, access control, and profile switching.

## What is an API Key Profile?

An **API Key Profile** lets you:
- **Map models**: Rewrite the model name from the client request
- **Restrict channels**: Limit the API Key to specific channels
- **Restrict models**: Limit the API Key to specific models
- **Switch profiles**: Create multiple profiles for one API Key and choose which one is active

In simple terms, an API Key Profile decides at the **request entry point** what model the request should be treated as first.

## Where API Key Profile Fits in the Request Flow

API Key Profile model mapping is the **first** step in a three-layer pipeline. For the full picture, see [Request Processing Guide](../getting-started/request-processing.md#core-concept-three-layers-of-model-settings).

In short: **API Key Profile renames → Model Association selects channel → Channel renames → Send upstream**

## Common Use Cases

### Use Case 1: Client tools with fixed model names

Many AI tools use fixed model names internally. If you want them to use other models, use API Key Profile model mapping.

```json
{
  "modelMappings": [
    {"from": "claude-sonnet-4-5", "to": "anthropic/claude-3.5-sonnet"}
  ]
}
```

### Use Case 2: Unify model names across clients

```json
{
  "modelMappings": [
    {"from": "gpt4", "to": "gpt-4o"},
    {"from": "gpt-4-turbo", "to": "gpt-4o"}
  ]
}
```

### Use Case 3: Restrict a profile to part of the system

```json
{
  "channelTags": ["production"],
  "modelIDs": ["gpt-4o", "claude-3-sonnet"]
}
```

## Override Model Channel Policies Per API Key

By default, requests use the global model associations configured in model management. You can also define independent model channel policies in an API Key Profile for specific models.

Use this when you want to:

- Route the same model to different providers for different customers or tools
- Give one API Key cheaper channels, fallback channels, or dedicated channels
- Keep the global model policy unchanged and only add exceptions for a few keys

Where to configure it:

1. Go to **API Keys**
2. Open **Profiles** for an API Key
3. Expand the profile you want to edit
4. Enter JSON in **Model Channel Policies**

### Basic Format

```json
[
  {
    "modelId": "gpt-4o",
    "associations": [
      {
        "type": "channel_model",
        "priority": 0,
        "channelModel": {
          "channelId": 12,
          "modelId": "gpt-4o"
        }
      }
    ]
  }
]
```

Meaning:

- `modelId`: the AxonHub model ID to override
- `associations`: the model association rules used by this API Key Profile
- `priority`: lower numbers have higher priority
- `channelId`: the channel ID
- `channelModel.modelId`: the model name sent to that channel

If **Model Channel Policies** is an empty array `[]` or is not configured, the API Key continues to use global model associations.

### Multiple Fallback Channels

```json
[
  {
    "modelId": "gpt-4o",
    "associations": [
      {
        "type": "channel_model",
        "priority": 0,
        "channelModel": {
          "channelId": 12,
          "modelId": "gpt-4o"
        }
      },
      {
        "type": "channel_model",
        "priority": 10,
        "channelModel": {
          "channelId": 18,
          "modelId": "gpt-4o"
        }
      }
    ]
  }
]
```

This prefers channel `12` and considers channel `18` during retry or failover.

### Select by Channel Tags

To avoid hardcoding channel IDs, use channel tags:

```json
[
  {
    "modelId": "gpt-4o-mini",
    "associations": [
      {
        "type": "channel_tags_model",
        "priority": 0,
        "channelTagsModel": {
          "channelTags": ["cheap"],
          "modelId": "gpt-4o-mini"
        }
      }
    ]
  }
]
```

This selects channels tagged with `cheap` that support `gpt-4o-mini`.

### Regex Match Channel Models

```json
[
  {
    "modelId": "gpt-4o",
    "associations": [
      {
        "type": "regex",
        "priority": 0,
        "regex": {
          "pattern": "gpt-4o.*",
          "exclude": []
        }
      }
    ]
  }
]
```

This matches models named `gpt-4o.*` across all channels.

### Important Notes

- `modelId` is the model name **after API Key Profile model mapping**.
  For example, if the profile maps `gpt-4` to `gpt-4o`, use `gpt-4o` here.
- The global model for `modelId` must be enabled. A disabled model cannot be revived by an API Key override.
- Once a key-level policy is configured for a `modelId`, that model no longer uses the global model association.
- If the key-level policy matches no channels, the request fails instead of falling back to the global policy.
- Project Profile and API Key Profile channel ID/tag restrictions still apply.

## Override Storage Policy Per API Key

By default, whether AxonHub stores request bodies, response bodies, stream chunks, and live previews is controlled by the global storage policy. You can enable **Storage Policy** in an API Key Profile to override those settings for one key.

Common use cases:

- Disable request or response body storage for external customers
- Enable stream chunk storage and live preview for internal debugging keys
- Send one API Key's payloads to a dedicated data storage backend

### Fields

```json
{
  "storagePolicy": {
    "dataStorageId": 3,
    "storeRequestBody": false,
    "storeResponseBody": true,
    "storeChunks": false,
    "livePreview": false
  }
}
```

Field meanings:

| Field | Description |
|-------|-------------|
| `dataStorageId` | Data Storage ID used for request and response bodies |
| `storeRequestBody` | Whether to store the client request body |
| `storeResponseBody` | Whether to store the upstream response body |
| `storeChunks` | Whether to store streaming response chunks |
| `livePreview` | Whether to enable live preview |

Every field is optional. Unset fields inherit the global storage policy.

In the UI, selecting **Inherit global** leaves that field unset. Only explicit enabled or disabled values override the global setting.

## Configuration Steps

### Step 1: Open the profile UI

1. Log in to the AxonHub management interface
2. Go to **API Keys**
3. Find the API Key you want to configure
4. Open the **Actions** menu
5. Select **Profiles** or **Configure**

### Step 2: Create a profile

1. Click **Add Profile**
2. Enter a profile name
3. Configure model mappings, channel restrictions, or model restrictions

### Step 3: Configure model mappings

Each mapping has:
- **From**: the model name in the client request
- **To**: the model name that AxonHub should use

Supported matching methods:

#### Exact match

```json
{"from": "gpt-4", "to": "claude-3-opus"}
```

#### Regex match

```json
{"from": "gpt-.*", "to": "claude-3-sonnet"}
```

### Step 4: Set the active profile

1. Select a profile in **Active Profile**
2. Click **Save**
3. The change takes effect immediately

## Rule Matching Order

Model mappings are evaluated in order. **The first matching rule is applied.**

Put more specific rules first and more general rules later.

## FAQ

### Q: Why is model mapping not working?

Check:
1. Is the correct active profile selected?
2. Does the model name match?
3. Is the regex pattern correct?

### Q: What are the requirements for profile names?

- Must be unique within one API Key
- Cannot be empty
- Should be meaningful

### Q: How many profiles should I create?

There is no hard limit, but keep the set small and easy to manage.

## Best Practices

1. **Use descriptive names** like `production` or `openrouter-mapping`
2. **Put specific rules first**
3. **Test before enabling**
4. **Prefer channel tags** over hardcoded channel IDs

## Related Documentation

- [Model Management Guide](model-management.md) - Configure model association
- [Channel Management Guide](channel-management.md) - Configure upstream channels
- [Load Balancing Guide](load-balance.md) - Understand channel selection and failover
- [Request Processing Guide](../getting-started/request-processing.md) - See the full request flow
