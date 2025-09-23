package eventbus

type EventBus interface {
	Publish(topic string, args ...interface{})
	Subscribe(topic string, fn interface{}) error
	Unsubscribe(topic string, fn interface{}) error
}
