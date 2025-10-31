package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/skr1ms/dev2gis/config"
	"github.com/skr1ms/dev2gis/pkg/middleware"
)

type Client struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  *config.RabbitMQConfig
	logger  middleware.LoggerInterface
}

func NewClient(cfg *config.RabbitMQConfig, logger middleware.LoggerInterface) (*Client, error) {
	var conn *amqp091.Connection
	var channel *amqp091.Channel
	var err error

	maxRetries := 30
	retryDelay := cfg.ReconnectDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		conn, err = amqp091.Dial(cfg.URL)
		if err != nil {
			if attempt < maxRetries {
				logger.Warn("Failed to connect to RabbitMQ, retrying...", map[string]interface{}{
					"attempt":     attempt,
					"max_retries": maxRetries,
					"retry_in":    retryDelay,
					"error":       err.Error(),
				})
				time.Sleep(retryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", maxRetries, err)
		}
		break
	}

	channel, err = conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	_, err = channel.QueueDeclare(
		cfg.QueueName,
		cfg.DurableQueue,
		cfg.AutoDelete,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	client := &Client{
		conn:    conn,
		channel: channel,
		config:  cfg,
		logger:  logger,
	}

	go client.handleReconnect()

	logger.Info("RabbitMQ client initialized successfully", map[string]interface{}{
		"queue": cfg.QueueName,
	})

	return client, nil
}

func (c *Client) PublishTask(ctx context.Context, task *HeightmapTask) error {
	if !c.IsConnected() || c.channel == nil || c.channel.IsClosed() {
		c.logger.Warn("RabbitMQ connection/channel is closed, attempting to reconnect before publishing")
		if err := c.reconnect(); err != nil {
			return fmt.Errorf("failed to reconnect before publishing: %w", err)
		}
	}

	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.config.PublishRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Second * time.Duration(attempt))
			c.logger.Warn("Retrying task publication", map[string]interface{}{
				"attempt": attempt,
				"job_id":  task.JobID,
			})

			if !c.IsConnected() || c.channel == nil || c.channel.IsClosed() {
				if err := c.reconnect(); err != nil {
					lastErr = err
					continue
				}
			}
		}

		err = c.channel.PublishWithContext(
			ctx,
			c.config.Exchange,
			c.config.QueueName,
			false,
			false,
			amqp091.Publishing{
				DeliveryMode: amqp091.Persistent,
				ContentType:  "application/json",
				Body:         body,
				Timestamp:    time.Now(),
				Priority:     uint8(task.Priority),
			},
		)

		if err == nil {
			c.logger.Info("Task published successfully", map[string]interface{}{
				"job_id": task.JobID,
				"queue":  c.config.QueueName,
			})
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("failed to publish task after %d retries: %w", c.config.PublishRetries, lastErr)
}

func (c *Client) PublishBatchTask(ctx context.Context, task *BatchHeightmapTask) error {
	if !c.IsConnected() || c.channel == nil || c.channel.IsClosed() {
		c.logger.Warn("RabbitMQ connection/channel is closed, attempting to reconnect before publishing batch task")
		if err := c.reconnect(); err != nil {
			return fmt.Errorf("failed to reconnect before publishing batch task: %w", err)
		}
	}

	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal batch task: %w", err)
	}

	batchQueueName := c.config.QueueName + "_batch"

	_, err = c.channel.QueueDeclare(
		batchQueueName,
		c.config.DurableQueue,
		c.config.AutoDelete,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare batch queue: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.config.PublishRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Second * time.Duration(attempt))
			c.logger.Warn("Retrying batch task publication", map[string]interface{}{
				"attempt":      attempt,
				"batch_job_id": task.BatchJobID,
			})

			if !c.IsConnected() || c.channel == nil || c.channel.IsClosed() {
				if err := c.reconnect(); err != nil {
					lastErr = err
					continue
				}
			}
		}

		err = c.channel.PublishWithContext(
			ctx,
			c.config.Exchange,
			batchQueueName,
			false,
			false,
			amqp091.Publishing{
				DeliveryMode: amqp091.Persistent,
				ContentType:  "application/json",
				Body:         body,
				Timestamp:    time.Now(),
				Priority:     uint8(task.Priority),
			},
		)

		if err == nil {
			c.logger.Info("Batch task published successfully", map[string]interface{}{
				"batch_job_id": task.BatchJobID,
				"queue":        batchQueueName,
				"image_count":  len(task.ImageURLs),
			})
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("failed to publish batch task after %d retries: %w", c.config.PublishRetries, lastErr)
}

func (c *Client) handleReconnect() {
	closeChan := make(chan *amqp091.Error)
	c.conn.NotifyClose(closeChan)

	for err := range closeChan {
		if err != nil {
			c.logger.Error("RabbitMQ connection closed, attempting to reconnect", err)

			for {
				time.Sleep(c.config.ReconnectDelay)

				conn, err := amqp091.Dial(c.config.URL)
				if err != nil {
					c.logger.Error("Failed to reconnect to RabbitMQ", err)
					continue
				}

				channel, err := conn.Channel()
				if err != nil {
					conn.Close()
					c.logger.Error("Failed to open channel on reconnect", err)
					continue
				}

				c.conn = conn
				c.channel = channel

				c.logger.Info("Successfully reconnected to RabbitMQ")

				closeChan = make(chan *amqp091.Error)
				c.conn.NotifyClose(closeChan)
				break
			}
		}
	}
}

func (c *Client) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}
	c.logger.Info("RabbitMQ client closed")
	return nil
}

func (c *Client) IsConnected() bool {
	return c.conn != nil && !c.conn.IsClosed()
}

func (c *Client) reconnect() error {
	c.logger.Info("Attempting to reconnect to RabbitMQ")

	conn, err := amqp091.Dial(c.config.URL)
	if err != nil {
		return fmt.Errorf("failed to dial RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	c.conn = conn
	c.channel = channel

	c.logger.Info("Successfully reconnected to RabbitMQ")

	go c.handleReconnect()

	return nil
}
