package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/config"
	"github.com/rabbitmq/amqp091-go"
)

type Rabbitmq struct {
	conn          *amqp091.Connection
	ch            *amqp091.Channel
	cfg           config.Rabbitmq
	notificationQ *amqp091.Queue
	clubQ         *amqp091.Queue
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

	err = ch.ExchangeDeclare(
		cfg.ExchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to declare exchange: %w", op, err)
	}

	nQ, err := ch.QueueDeclare(
		"notification",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to declare notification queue: %w", op, err)
	}

	cQ, err := ch.QueueDeclare(
		"club",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to declare notification queue: %w", op, err)
	}

	err = ch.QueueBind(
		nQ.Name,
		"user.notification.*",
		cfg.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to bind exchange to queue: %w", op, err)
	}

	err = ch.QueueBind(
		cQ.Name,
		"user.club.*",
		cfg.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to bind exchange to queue: %w", op, err)
	}

	return &Rabbitmq{
		conn:          conn,
		ch:            ch,
		notificationQ: &nQ,
		clubQ:         &cQ,
		cfg:           cfg,
	}, nil
}

func (r *Rabbitmq) Publish(ctx context.Context, routingKey string, msg any) error {
	const op = "Rabbitmq.Publish"

	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = r.ch.PublishWithContext(
		ctx,
		r.cfg.ExchangeName,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent,
			ContentType:  "application/json",
			Body:         bytes,
		})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
