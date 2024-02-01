package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/config"
	"github.com/rabbitmq/amqp091-go"
)

type Rabbitmq struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
	q    *amqp091.Queue
}

func New(cfg config.Rabbitmq) (*Rabbitmq, error) {
	const op = "Rabbitmq.New"

	connString := fmt.Sprintf("amqp://%v:%v@%v:%v/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	conn, err := amqp091.Dial(connString)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to amqp server: %w", op, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open a channel: %w", op, err)
	}

	q, err := ch.QueueDeclare(
		cfg.UserActivationEmailQueue.Name,
		cfg.UserActivationEmailQueue.Durable,
		cfg.UserActivationEmailQueue.AutoDelete,
		cfg.UserActivationEmailQueue.Exclusive,
		cfg.UserActivationEmailQueue.NoWait,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to declare a queue: %w", op, err)
	}

	return &Rabbitmq{
		conn: conn,
		ch:   ch,
		q:    &q,
	}, nil
}

func (r *Rabbitmq) Publish(ctx context.Context, msg any) error {
	const op = "Rabbitmq.Publish"

	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = r.ch.PublishWithContext(ctx,
		"",
		r.q.Name,
		false,
		false,
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent,
			ContentType:  "text/plain",
			Body:         bytes,
		})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
