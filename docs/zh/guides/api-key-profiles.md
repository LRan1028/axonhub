# API Key Profile 指南

本文介绍如何配置 API Key Profile，实现模型映射、访问控制和多 Profile 切换。

## 什么是 API Key Profile？

**API Key Profile** 让你可以：
- **模型映射**：把客户端请求的模型名改成另一个模型
- **渠道限制**：限制 API Key 只能使用特定渠道
- **模型限制**：限制 API Key 只能访问特定模型
- **多 Profile 切换**：为同一个 API Key 创建多个 Profile，并选择当前生效的 Profile

通俗地说：API Key Profile 是在请求**入口处**决定“这个请求先按什么模型处理”。

## API Key Profile 在请求流程中的位置

API Key Profile 的模型映射是三层流水线中的**第一步**。完整说明请参阅 [请求处理流程](../getting-started/request-processing.md#核心概念三层模型设置)。

简单来说：**API Key Profile 改模型名 → 模型关联选渠道 → 渠道改模型名 → 发给上游**

## 模型映射的使用场景

### 场景 1：客户端工具使用固定模型名

很多 AI 工具会在内部使用固定模型名。如果你想让它们实际走别的模型，就需要 API Key Profile 模型映射。

```json
{
  "modelMappings": [
    {"from": "claude-sonnet-4-5", "to": "anthropic/claude-3.5-sonnet"}
  ]
}
```

### 场景 2：统一不同客户端的模型名称

```json
{
  "modelMappings": [
    {"from": "gpt4", "to": "gpt-4o"},
    {"from": "gpt-4-turbo", "to": "gpt-4o"}
  ]
}
```

### 场景 3：限制 Profile 的可用范围

```json
{
  "channelTags": ["production"],
  "modelIDs": ["gpt-4o", "claude-3-sonnet"]
}
```

## 按 API Key 覆盖模型渠道策略

默认情况下，请求会使用全局模型管理里的“模型关联”来选择渠道。你也可以在某个 API Key Profile 里为指定模型配置独立的模型渠道策略。

这适合这些场景：

- 同一个模型，对不同客户或不同工具走不同供应商
- 给某个 API Key 单独指定低价渠道、备用渠道或专属渠道
- 保持全局模型策略不变，只对少数 Key 做例外

配置位置：

1. 进入 **API Keys**
2. 打开某个 API Key 的 **Profiles**
3. 展开要配置的 Profile
4. 在 **模型渠道策略** 中填写 JSON

### 基本格式

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

含义：

- `modelId`：要覆盖的 AxonHub 模型 ID
- `associations`：这个模型在当前 API Key Profile 下使用的模型关联规则
- `priority`：优先级，数值越小越优先
- `channelId`：渠道 ID
- `channelModel.modelId`：发给该渠道时使用的渠道模型名

如果 **模型渠道策略** 为空数组 `[]` 或未配置，该 API Key 会继续使用全局模型关联。

### 多个备用渠道

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

上面的配置会优先使用渠道 `12`，失败或重试切换时再考虑渠道 `18`。

### 按渠道标签选择

如果不想硬编码渠道 ID，可以使用渠道标签：

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

这会从带有 `cheap` 标签的渠道中选择支持 `gpt-4o-mini` 的渠道。

### 正则匹配渠道模型

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

这会匹配所有渠道里符合 `gpt-4o.*` 的模型。

### 注意事项

- `modelId` 写的是 **API Key Profile 模型映射之后** 的模型名。
  例如 Profile 先把 `gpt-4` 映射成 `gpt-4o`，这里就要写 `gpt-4o`。
- `modelId` 对应的全局模型必须处于启用状态。禁用模型不会因为 API Key 覆盖策略而重新可用。
- 一旦某个 `modelId` 配置了 key 级策略，该模型就不会再使用全局模型关联。
- 如果 key 级策略没有匹配到任何渠道，请求会失败，不会自动回退到全局策略。
- 仍然会继续受 Project Profile 和 API Key Profile 的渠道 ID、渠道标签限制影响。

## 按 API Key 覆盖存储策略

默认情况下，请求体、响应体、流式块和实时预览是否保存，由系统全局存储策略决定。你可以在 API Key Profile 中打开 **存储策略**，让某个 Key 使用独立配置。

常见场景：

- 对外部客户禁用请求体或响应体保存，减少敏感数据留存
- 对内部调试 Key 开启流式块保存和实时预览
- 给某个 Key 指定独立的数据存储后端

### 配置字段

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

字段含义：

| 字段 | 说明 |
|------|------|
| `dataStorageId` | 指定请求和响应正文使用的数据存储 ID |
| `storeRequestBody` | 是否保存客户端请求体 |
| `storeResponseBody` | 是否保存上游响应体 |
| `storeChunks` | 是否保存流式响应块 |
| `livePreview` | 是否启用实时预览 |

每个字段都可以不设置。不设置表示继承全局存储策略。

如果在界面中选择 **继承全局**，保存时对应字段会保持为空。只有明确选择启用或禁用时，才会覆盖全局设置。

## 配置步骤

### 步骤 1：进入配置界面

1. 登录 AxonHub 管理界面
2. 进入 **API Keys** 页面
3. 找到要配置的 API Key
4. 点击右侧的 **操作** 菜单
5. 选择 **Profiles** 或 **配置**

### 步骤 2：创建 Profile

1. 点击 **新增配置**
2. 输入 Profile 名称
3. 配置模型映射、渠道限制或模型限制

### 步骤 3：配置模型映射

每个映射包含：
- **From（源模型）**：客户端请求的模型名称
- **To（目标模型）**：实际使用的模型名称

支持两种匹配方式：

#### 精确匹配

```json
{"from": "gpt-4", "to": "claude-3-opus"}
```

#### 正则匹配

```json
{"from": "gpt-.*", "to": "claude-3-sonnet"}
```

### 步骤 4：设置生效 Profile

1. 在 **生效配置** 下拉菜单中选择要使用的 Profile
2. 点击 **保存**
3. 配置立即生效

## 规则匹配顺序

模型映射按顺序匹配，**第一个匹配的规则会生效**。

建议把更具体的规则放前面，把更通用的规则放后面。

## 常见问题

### Q: 模型映射不生效？

检查：
1. 是否选择了正确的生效 Profile
2. 模型名称是否匹配
3. 正则表达式是否正确

### Q: Profile 名称有什么要求？

- 在同一个 API Key 内必须唯一
- 不能为空
- 建议使用有意义的名称

### Q: 可以创建多少个 Profile？

没有硬性限制，但建议保持简洁，便于管理。

## 最佳实践

1. **使用描述性名称**：如 `production`、`openrouter-mapping`
2. **具体规则在前**：把精确映射放前面，通用映射放后面
3. **先测试再启用**：先验证映射结果是否符合预期
4. **优先使用渠道标签**：比硬编码渠道 ID 更灵活

## 相关文档

- [模型管理指南](model-management.md) - 配置模型关联
- [渠道配置指南](channel-management.md) - 配置上游渠道
- [负载均衡指南](load-balance.md) - 了解渠道选择和故障转移
- [请求处理流程](../getting-started/request-processing.md) - 了解完整请求链路
