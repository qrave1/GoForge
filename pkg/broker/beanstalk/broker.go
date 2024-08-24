package beanstalk

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ikool-cn/gobeanstalk-connection-pool"
	"github.com/qrave1/GoForge/pkg/broker"
)

const (
	defaultTimeout = 10 * time.Second
)

type BrokerEngine struct {
	mu         sync.Mutex
	bnstlkPool *gobeanstalk.Pool
	subs       []*subscriber
	stop       chan struct{}
}

func NewBrokerEngine(addr string) (broker.Broker, error) {
	return &BrokerEngine{
		bnstlkPool: &gobeanstalk.Pool{
			Dial: func() (*gobeanstalk.Conn, error) {
				conn, err := gobeanstalk.Dial(addr)
				if err != nil {
					return nil, fmt.Errorf("error dial beanstalkd: %w", err)
				}

				return conn, nil
			},
		},
	}, nil
}

func (b *BrokerEngine) Publish(topic string, event broker.Message) error {
	conn, err := b.bnstlkPool.Get()
	if err != nil {
		return fmt.Errorf("get beanstalk pool conn: %w", err)
	}

	defer func() {
		_ = b.bnstlkPool.Release(conn, err != nil)
	}()

	if err = conn.Use(topic); err != nil {
		return fmt.Errorf("use beanstalk topic: %w", err)
	}

	if _, err = conn.Put(event.Message(), 0, 0, defaultTimeout); err != nil {
		return fmt.Errorf("put beanstalk event: %w", err)
	}

	return nil
}

func (b *BrokerEngine) Subscribe(topic string, handler broker.HandlerFunc) error {
	conn, err := b.bnstlkPool.Get()
	if err != nil {
		return fmt.Errorf("get beanstalk pool conn: %w", err)
	}

	sub := &subscriber{
		topic:  topic,
		handle: handler,
		stop:   b.stop,
	}

	go sub.startWorker(conn)

	b.addSubscriber(sub)

	return nil
}

func (b *BrokerEngine) addSubscriber(sub *subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs = append(b.subs, sub)
}

func (b *BrokerEngine) Close() error {
	close(b.stop)
	return b.bnstlkPool.Close()
}

type subscriber struct {
	topic  string
	handle broker.HandlerFunc
	stop   chan struct{}
}

func (s *subscriber) startWorker(conn *gobeanstalk.Conn) {
	for {
		select {
		case <-s.stop:
			return
		default:
			if err := s.work(conn); err != nil && !errors.Is(err, gobeanstalk.ErrTimedOut) {
				return // todo мб и не возвращать при ошибке
			}
		}
	}
}

func (s *subscriber) work(conn *gobeanstalk.Conn) error {
	job, err := conn.Reserve(defaultTimeout)
	if err != nil {
		return fmt.Errorf("reserve job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("job is nil")
	}

	if err = s.handle(
		&beanstalkMessage{
			topic:   s.topic,
			payload: job.Body,
		},
	); err != nil {
		_ = conn.Delete(job.ID)
		return fmt.Errorf("error handle beanstalkd event: %w", err)
	}

	if err = conn.Delete(job.ID); err != nil {
		return fmt.Errorf("error delete job by id:%d :%w", job.ID, err)
	}

	return nil
}
