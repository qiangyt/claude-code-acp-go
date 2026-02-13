package acp

// 本文件从官方 ACP Go SDK 导出类型，方便项目内使用
// 官方 SDK: github.com/coder/acp-go-sdk

import (
	acp "github.com/coder/acp-go-sdk"
)

// ============================================================
// 协议核心类型 - 从官方 SDK 重新导出
// ============================================================

// 协议版本
type ProtocolVersion = acp.ProtocolVersion

// 请求 ID
type RequestId = acp.RequestId

// ============================================================
// 能力类型
// ============================================================

// 实现信息
type Implementation = acp.Implementation

// 文件系统能力
type FileSystemCapability = acp.FileSystemCapability

// 客户端能力
type ClientCapabilities = acp.ClientCapabilities

// Prompt 能力
type PromptCapabilities = acp.PromptCapabilities

// MCP 能力
type McpCapabilities = acp.McpCapabilities

// 会话能力
type SessionCapabilities = acp.SessionCapabilities

// 代理能力
type AgentCapabilities = acp.AgentCapabilities

// ============================================================
// 认证类型
// ============================================================

// 认证方法
type AuthMethod = acp.AuthMethod

// ============================================================
// 初始化类型
// ============================================================

// 初始化请求
type InitializeRequest = acp.InitializeRequest

// 初始化响应
type InitializeResponse = acp.InitializeResponse

// ============================================================
// 内容块类型
// ============================================================

// 内容块（联合类型）
type ContentBlock = acp.ContentBlock

// 文本内容块
type ContentBlockText = acp.ContentBlockText

// 图片内容块
type ContentBlockImage = acp.ContentBlockImage

// 音频内容块
type ContentBlockAudio = acp.ContentBlockAudio

// 资源链接内容块
type ContentBlockResourceLink = acp.ContentBlockResourceLink

// 嵌入式资源内容块
type ContentBlockResource = acp.ContentBlockResource

// 图片内容
type ImageContent = acp.ImageContent

// 音频内容
type AudioContent = acp.AudioContent

// 嵌入式资源
type EmbeddedResource = acp.EmbeddedResource

// 资源链接
type ResourceLink = acp.ResourceLink

// ============================================================
// 会话类型
// ============================================================

// 会话 ID
type SessionId = acp.SessionId

// 新会话请求
type NewSessionRequest = acp.NewSessionRequest

// 新会话响应
type NewSessionResponse = acp.NewSessionResponse

// 加载会话请求
type LoadSessionRequest = acp.LoadSessionRequest

// 加载会话响应
type LoadSessionResponse = acp.LoadSessionResponse

// ============================================================
// 会话更新类型
// ============================================================

// 会话更新（联合类型）
type SessionUpdate = acp.SessionUpdate

// 用户消息块更新
type SessionUpdateUserMessageChunk = acp.SessionUpdateUserMessageChunk

// 代理消息块更新
type SessionUpdateAgentMessageChunk = acp.SessionUpdateAgentMessageChunk

// 代理思考块更新
type SessionUpdateAgentThoughtChunk = acp.SessionUpdateAgentThoughtChunk

// 工具调用更新
type SessionUpdateToolCall = acp.SessionUpdateToolCall

// 工具调用状态更新
type SessionToolCallUpdate = acp.SessionToolCallUpdate

// 计划更新
type SessionUpdatePlan = acp.SessionUpdatePlan

// ============================================================
// 工具调用类型
// ============================================================

// 工具调用 ID
type ToolCallId = acp.ToolCallId

// 工具调用状态
type ToolCallStatus = acp.ToolCallStatus

// 工具类型
type ToolKind = acp.ToolKind

// 工具调用内容
type ToolCallContent = acp.ToolCallContent

// 工具调用位置
type ToolCallLocation = acp.ToolCallLocation

// ============================================================
// 计划类型
// ============================================================

// 计划条目
type PlanEntry = acp.PlanEntry

// ============================================================
// 停止原因
// ============================================================

// 停止原因
type StopReason = acp.StopReason

// ============================================================
// 常量
// ============================================================

// 工具类型常量
const (
	ToolKindRead       = acp.ToolKindRead
	ToolKindEdit       = acp.ToolKindEdit
	ToolKindDelete     = acp.ToolKindDelete
	ToolKindMove       = acp.ToolKindMove
	ToolKindSearch     = acp.ToolKindSearch
	ToolKindExecute    = acp.ToolKindExecute
	ToolKindThink      = acp.ToolKindThink
	ToolKindFetch      = acp.ToolKindFetch
	ToolKindSwitchMode = acp.ToolKindSwitchMode
	ToolKindOther      = acp.ToolKindOther
)

// 工具调用状态常量
const (
	ToolCallStatusPending    = acp.ToolCallStatusPending
	ToolCallStatusInProgress = acp.ToolCallStatusInProgress
	ToolCallStatusCompleted  = acp.ToolCallStatusCompleted
	ToolCallStatusFailed     = acp.ToolCallStatusFailed
)

// 停止原因常量
const (
	StopReasonEndTurn         = acp.StopReasonEndTurn
	StopReasonMaxTokens       = acp.StopReasonMaxTokens
	StopReasonMaxTurnRequests = acp.StopReasonMaxTurnRequests
	StopReasonRefusal         = acp.StopReasonRefusal
	StopReasonCancelled       = acp.StopReasonCancelled
)
