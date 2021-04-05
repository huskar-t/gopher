package mq


type Subscriber interface {
	Unsubscribe() error
}

type MQ interface {
	Stop()
	Producer
	Consumer
}

type Producer interface {
	Publish(topic string, data interface{}) error
}

type Consumer interface {
	Subscribe(topic string, cb CallBack) (Subscriber, error)
	GroupSubscribe(topic, group string, cb CallBack) (Subscriber, error)
}

type CallBack func(topic string,message interface{})