package queuetool

import (
	properties "CUGOj-Data/src/Properties"

	"github.com/streadway/amqp"
)

func RabbitMQConn() (*amqp.Connection, error) {
	user, err := properties.Get("QuUser")
	if err != nil {
		return nil, err
	}
	pwd, err := properties.Get("QuPassword")
	if err != nil {
		return nil, err
	}
	host, err := properties.Get("QuHost")
	if err != nil {
		return nil, err
	}
	port, err := properties.Get("QuPort")
	if err != nil {
		return nil, err
	}
	url := "amqp://" + user + ":" + pwd + "@" + host + ":" + port + "/"
	conn, err := amqp.Dial(url)
	return conn, err
}

func NewWork(qu, msg string) error {
	conn, err := RabbitMQConn()
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(
		qu,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	err = ch.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return err
	}
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(msg),
		},
	)
	return err
}
