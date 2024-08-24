package broker

import (
	"io"
)

// Publisher описывает интерфейс для публикации сообщений в брокер
type Publisher interface {
	// Publish отправляет событие в указанный топик или очередь
	Publish(topic string, msg Message) error
}

// Subscriber описывает интерфейс для подписки и получения сообщений из брокера
type Subscriber interface {
	// Subscribe подписывается на указанный топик или очередь и обрабатывает события с помощью переданной функции
	Subscribe(topic string, handler HandlerFunc) error
}

// HandlerFunc функция-обработчик для сообщения
type HandlerFunc func(msg Message) error

// Broker описывает интерфейс для брокера сообщений
type Broker interface {
	Publisher
	Subscriber
	io.Closer
}
