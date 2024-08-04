package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ExchangeName string
type RoutingKey string

const (
	UserExchangeName ExchangeName = "user-exchange"

	PushNotification   RoutingKey = "user.notification.push"
	UserRegistered     RoutingKey = "user.notification.email.registered"
	UserUpdated        RoutingKey = "user.event.updated"
	UserActivated      RoutingKey = "user.event.activated"
	UserDeleted        RoutingKey = "user.event.deleted"
	UserForgotPassword RoutingKey = "user.notification.email.forgot-password"
)

type Rabbitmq struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	cfg  config.Rabbitmq
}

func (r *Rabbitmq) Close() error {
	const op = "rabbitmq.close"

	err := r.ch.Close()
	if err != nil {
		return fmt.Errorf("%s: failed to close channel: %w", op, err)
	}

	err = r.conn.Close()
	if err != nil {
		return fmt.Errorf("%s: failed to close connection: %w", op, err)
	}

	return nil
}

func New(cfg config.Rabbitmq) (*Rabbitmq, error) {
	const op = "Rabbitmq.New"

	connString := fmt.Sprintf("amqp://%v:%v@%v:%v/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	conn, err := amqp.Dial(connString)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to amqp server: %w", op, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open a channel: %w", op, err)
	}

	err = ch.ExchangeDeclare(
		string(UserExchangeName),
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

	return &Rabbitmq{
		conn: conn,
		ch:   ch,
		cfg:  cfg,
	}, nil
}

func (r *Rabbitmq) Publish(ctx context.Context, exchangeName ExchangeName, routingKey RoutingKey, msg any) error {
	const op = "Rabbitmq.Publish"

	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = r.ch.PublishWithContext(
		ctx,
		string(exchangeName),
		string(routingKey),
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         bytes,
		})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
