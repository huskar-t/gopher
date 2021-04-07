package nats

import (
	"errors"
	"github.com/huskar-t/gopher/common/define/mq"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type Nats struct {
	nc     *nats.Conn
	ec     *nats.EncodedConn
	logger logrus.FieldLogger
}

func (mq *Nats) Connect(conf *Config) error {
	url := conf.Addr
	if url == "" {
		url = nats.DefaultURL
	}

	opts := []nats.Option{
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			mq.logger.Infof("Got disconnected! Reason: %q\n", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			mq.logger.Infof("Got reconnected to %v!\n", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			mq.logger.Infof("Connection closed. Reason: %q\n", nc.LastError())
		}),
	}
	if conf.MaxReconnects > 0 {
		opts = append(opts, nats.MaxReconnects(conf.MaxReconnects))
	}
	if conf.ReconnectWait > 0 {
		opts = append(opts, nats.ReconnectWait(time.Duration(conf.ReconnectWait)*time.Second))
	}
	if conf.Token != "" {
		opts = append(opts, nats.Token(conf.Token))
	}
	if conf.Username != "" {
		opts = append(opts, nats.UserInfo(conf.Username, conf.Password))
	}
	var err error
	mq.nc, err = nats.Connect(url, opts...)
	if err != nil {
		return err
	}
	mq.ec, err = nats.NewEncodedConn(mq.nc, JSON_ENCODER)
	if err != nil {
		return err
	}
	return nil
}

func (mq *Nats) Stop() {
	if mq.ec != nil {
		mq.ec.Close()
	}
}

func (mq *Nats) Publish(topic string, data interface{}) error {
	if mq.ec == nil {
		return errors.New("broker not connected")
	}
	return mq.ec.Publish(topic, data)
}

func (mq *Nats) Subscribe(topic string, fn mq.CallBack) (mq.Subscriber, error) {
	if mq.ec == nil {
		return nil, errors.New("nats not connected")
	}
	topic = changeTopic(topic)
	return mq.ec.Subscribe(topic, fn)
}

func (mq *Nats) GroupSubscribe(topic, group string, fn mq.CallBack) (mq.Subscriber, error) {
	if mq.ec == nil {
		return nil, errors.New("nats not connected")
	}
	topic = changeTopic(topic)
	return mq.ec.QueueSubscribe(topic, group, fn)
}

func changeTopic(topic string) string {
	if strings.HasSuffix(topic, ".*") {
		return strings.TrimSuffix(topic, ".*") + ".>"
	} else if strings.HasSuffix(topic, "*") {
		return strings.TrimSuffix(topic, "*") + ".>"
	}
	return topic
}

func NewNatsMQ(conf *Config, logger logrus.FieldLogger) mq.MQ {
	natsMQ := &Nats{
		logger: logger,
	}
	reconnectWait := conf.ReconnectWait
	if reconnectWait <= 0 {
		reconnectWait = 2
	}
	var connected = make(chan bool)
	go func() {
		for {
			err := natsMQ.Connect(conf)
			if err != nil {
				logger.WithError(err).Error("connect to mq broker error")
				time.Sleep(time.Duration(reconnectWait) * time.Second)
			} else {
				break
			}
		}
		connected <- true
		logger.Info("nats server connected")
	}()
	<-connected
	return natsMQ
}
