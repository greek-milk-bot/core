package models

type Plugin interface {
	// Boot 插件协商绑定和初始化
	Boot(ctx PluginBus) error
}

type MessageReceiver interface {
	// ReceiveMessage 接收消息
	ReceiveMessage(ctx PluginBus, msg WithSrcPacket[Message]) error
}

type EventReceiver interface {
	// ReceiveEvent 接收消息
	ReceiveEvent(ctx PluginBus, msg WithSrcPacket[Event]) error
}
