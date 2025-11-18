package accrual

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ArtShib/gophermart.git/internal/config"
	"github.com/ArtShib/gophermart.git/internal/models"
)

type StoreOrder interface {
	GetOrdersInWork(ctx context.Context) (models.OrderArray, error)
	UpdateOrdersBatch(ctx context.Context, orders models.ResAccrualOrderArray) error
}

type HTTPClient interface {
	RequestAccrualOrder(ctx context.Context, urlConnect string) (*models.ResAccrualOrder, error)
}
type ClientAccrual struct {
	log           *slog.Logger
	store         StoreOrder
	client        HTTPClient
	urlConnect    string
	chListOrders  chan *models.Order
	chListAccrual chan *models.ResAccrualOrder
	buffer        models.ResAccrualOrderArray
	activeWorkers int32
	wg            sync.WaitGroup
	mu            sync.Mutex
	cancel        context.CancelFunc
	config        config.WorkerConfig
}

func New(log *slog.Logger, store StoreOrder, cfg config.WorkerConfig, client HTTPClient, urlConnect string) *ClientAccrual {
	return &ClientAccrual{
		log:           log,
		store:         store,
		chListOrders:  make(chan *models.Order, cfg.CountWorkers),
		chListAccrual: make(chan *models.ResAccrualOrder, cfg.InputChainSize),
		buffer:        make(models.ResAccrualOrderArray, 0, cfg.BufferSize),
		config:        cfg,
		client:        client,
		urlConnect:    urlConnect,
	}
}

func (c *ClientAccrual) Start(ctx context.Context) {
	c.log.Info("Sart Pool")
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	go c.GeneratorListOrders(ctx, c.chListOrders)
	go c.scaleWorkers(ctx)

	c.wg.Add(2)
	go c.UpdateOrders(ctx)

}

func (c *ClientAccrual) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
}

func (c *ClientAccrual) GeneratorListOrders(ctx context.Context, chListOrders chan *models.Order) {
	defer c.wg.Done()
	const op = "Accrual.GeneratorListOrders"

	log := c.log.With(
		slog.String("op", op))
	log.Info("start generator list orders")

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			orders, err := c.store.GetOrdersInWork(ctx)
			if err != nil {
				if errors.Is(err, models.ErrOrdersInWorkIsEmpty) {
					c.log.Info("list of orders is empty", "info", err)
					continue
				}
				c.log.Error("failed to get orders", "error", err)
				continue
			}
			for _, order := range orders {
				chListOrders <- &order
			}
		case <-ctx.Done():
			close(chListOrders)
			log.Info("ctx.Done")
			return
		}
	}
}

func (c *ClientAccrual) scaleWorkers(ctx context.Context) {
	const op = "Accrual.scaleWorkers"

	log := c.log.With(
		slog.String("op", op))
	log.Info("start scale workers")

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			queueLen := len(c.chListOrders)
			activeWorkers := atomic.LoadInt32(&c.activeWorkers)

			if queueLen > 0 && activeWorkers < c.config.CountWorkers {
				c.wg.Add(1)
				atomic.AddInt32(&c.activeWorkers, 1)
				go c.worker(ctx, int(activeWorkers)+1, c.chListOrders)
				c.log.Info("Add Worker", "ID", atomic.LoadInt32(&c.activeWorkers))
			}
		}
	}
}

func (c *ClientAccrual) worker(ctx context.Context, id int, chOrder <-chan *models.Order) {
	const op = "Accrual.worker"

	log := c.log.With(
		slog.String("op", op),
		slog.String("id worker", fmt.Sprint(id)),
	)

	log.Info("start worker")

	defer func() {
		c.wg.Done()
		atomic.AddInt32(&c.activeWorkers, -1)

	}()
	for {
		select {
		case <-ctx.Done():
			return
		case order, ok := <-chOrder:
			if !ok {
				return
			}
			url := fmt.Sprintf("%s/api/orders/%d", c.urlConnect, order.Number)
			orderAccrual, err := c.client.RequestAccrualOrder(ctx, url)
			if err != nil {
				c.log.Error("failed to request accrual order", "error", err)
				continue //return
			}
			if orderAccrual.Status != order.Status || orderAccrual.Accrual != order.Accrual {
				c.chListAccrual <- orderAccrual
			}
		}
	}
}

func (c *ClientAccrual) UpdateOrders(ctx context.Context) {
	defer c.wg.Done()

	const op = "Accrual.UpdateOrders"

	log := c.log.With(
		slog.String("op", op))
	log.Info("start update Orders")

	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.flushBuffer(ctx)
			return

		case task := <-c.chListAccrual:
			c.addToBuffer(ctx, task)

		case <-ticker.C:
			c.flushIfNeeded(ctx)
		}
	}
}

func (c *ClientAccrual) addToBuffer(ctx context.Context, task *models.ResAccrualOrder) {
	const op = "Accrual.addToBuffer"

	log := c.log.With(
		slog.String("op", op))
	log.Info("start addToBuffer")

	c.mu.Lock()
	defer c.mu.Unlock()

	c.buffer = append(c.buffer, task)

	if len(c.buffer) >= c.config.BatchSize {
		go c.flushBufferAsync(ctx, c.getBufferCopy())
		c.buffer = c.buffer[:0]
	}
}

func (c *ClientAccrual) flushBuffer(ctx context.Context) {
	const op = "Accrual.flushBuffer"

	//log := c.log.With(
	//	slog.String("op", op))
	//log.Info("start flushBuffer")

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.buffer) > 0 {
		batch := make(models.ResAccrualOrderArray, len(c.buffer))
		copy(batch, c.buffer)
		c.buffer = c.buffer[:0]

		c.log.Info("Flushing buffer on shutdown",
			"batch_size", len(batch),
		)
		c.processBatch(ctx, batch)
	}
}

func (c *ClientAccrual) flushIfNeeded(ctx context.Context) {
	const op = "Accrual.flushIfNeeded"

	//log := c.log.With(
	//	slog.String("op", op))
	//log.Info("start flushIfNeeded")

	c.mu.Lock()
	if len(c.buffer) == 0 {
		c.mu.Unlock()
		return
	}

	batch := c.getBufferCopy()
	c.buffer = c.buffer[:0]
	c.mu.Unlock()

	c.processBatch(ctx, batch)
}

func (c *ClientAccrual) getBufferCopy() models.ResAccrualOrderArray {
	const op = "Accrual.getBufferCopy"

	//log := c.log.With(
	//	slog.String("op", op))
	//log.Info("start getBufferCopy")

	batch := make(models.ResAccrualOrderArray, len(c.buffer))
	copy(batch, c.buffer)
	return batch
}

func (c *ClientAccrual) flushBufferAsync(ctx context.Context, batch models.ResAccrualOrderArray) {
	const op = "Accrual.flushBufferAsync"

	log := c.log.With(
		slog.String("op", op))
	log.Info("start flushBufferAsync")

	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	c.processBatch(ctx, batch)
}

func (c *ClientAccrual) processBatch(ctx context.Context, batch models.ResAccrualOrderArray) {
	const op = "Accrual.processBatch"

	log := c.log.With(
		slog.String("op", op))
	log.Info("start processBatch - UpdateOrdersBatch")
	if err := c.store.UpdateOrdersBatch(ctx, batch); err != nil {
		c.log.Error("Batch processing failed",
			"error", err,
			"batch_size", len(batch))
	} else {
		c.log.Info("Batch processed successfully",
			"batch_size", len(batch))
	}
}
