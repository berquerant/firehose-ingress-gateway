package amqp

import (
	"fmt"

	queue "github.com/streadway/amqp"
)

//go:generate go run github.com/berquerant/goconfig@v0.2.0 -field "ExchangeName string|ExchangeKind string" -prefix Conn -option -output conn_conn_config_generated.go

type Conn struct {
	conn    *queue.Connection
	channel *queue.Channel
	config  *ConnConfig
}

func NewConn(conn *queue.Connection, opt ...ConnConfigOption) *Conn {
	config := NewConnConfigBuilder().Build()
	config.Apply(opt...)
	return &Conn{
		conn:   conn,
		config: config,
	}
}

func (c *Conn) Close() error {
	return c.channel.Close()
}

func (c *Conn) Init() error {
	var err error
	if c.channel, err = c.conn.Channel(); err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	if err := c.channel.ExchangeDeclare(
		c.config.ExchangeName.Get(),
		c.config.ExchangeKind.Get(),
		false, // durable
		false, // auto delete
		false, // internal
		false, // no wait
		nil,   // args
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}
	return nil
}

//go:generate go run github.com/berquerant/goconfig@v0.2.0 -field "QueueName string|ConsumerName string|PrefetchCount int" -prefix Consumer -option -output conn_consumer_config_generated.go

type Consumer struct {
	conn      *Conn
	config    *ConsumerConfig
	deliveryC <-chan queue.Delivery
}

func NewConsumer(conn *Conn, opt ...ConsumerConfigOption) *Consumer {
	config := NewConsumerConfigBuilder().Build()
	config.Apply(opt...)
	return &Consumer{
		conn:   conn,
		config: config,
	}
}

func (c *Consumer) Init() error {
	if _, err := c.conn.channel.QueueDeclare(
		c.config.QueueName.Get(),
		true,  // durable
		false, // auto delete
		false, // exclusive
		false, // no wait
		nil,   // args
	); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	if err := c.conn.channel.QueueBind(
		c.config.QueueName.Get(),
		"#", // match any
		c.conn.config.ExchangeName.Get(),
		false, // no wait
		nil,   // args
	); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	if err := c.conn.channel.Qos(c.config.PrefetchCount.Get(), 0, false); err != nil {
		return fmt.Errorf("failed to apply qos: %w", err)
	}

	deliveryC, err := c.conn.channel.Consume(
		c.config.QueueName.Get(),
		c.config.ConsumerName.Get(),
		false, // auto ack
		false, // exclusive
		false, // nonlocal
		false, // no wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to consume: %w", err)
	}
	c.deliveryC = deliveryC
	return nil
}

func (c *Consumer) Channel() <-chan queue.Delivery {
	return c.deliveryC
}

func (c *Consumer) Close() error {
	if err := c.conn.channel.Cancel(c.config.ConsumerName.Get(), false); err != nil {
		return fmt.Errorf("consumer failed to close channel: %w", err)
	}
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("consumer failed to close conn: %w", err)
	}
	return nil
}
