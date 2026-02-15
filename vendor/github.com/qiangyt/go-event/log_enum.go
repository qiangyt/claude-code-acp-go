package event

type LogEnum int64

const (
	_ LogEnum = iota
	HubCloseBegin
	HubCloseOk

	ListenerSubOk
	ListenerSubErr
	ListenerUnsubOk
	ListenerUnsubErr

	ListenerCloseBegin
	ListenerCloseOk

	TopicRegisterBegin
	TopicRegisterOk

	TopicCloseBegin
	TopicCloseOk

	EventPubBegin
	EventPubError
	EventPubOk

	EventSendBegin
	EventSendOk

	EventHandleBegin
	EventHandleOk
	EventHandleErr
)
