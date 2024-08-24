package beanstalk

type beanstalkMessage struct {
	topic   string
	payload []byte
}

func (e beanstalkMessage) Topic() string {
	return e.topic
}

func (e beanstalkMessage) Message() []byte {
	return e.payload
}
