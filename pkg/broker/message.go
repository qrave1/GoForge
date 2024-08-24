package broker

// Message представляет абстрактное сообщение, которое передаётся через брокер
type Message interface {
	Topic() string
	Message() []byte
}
