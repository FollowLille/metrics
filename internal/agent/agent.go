// Package agent содержит логику работы агента
// Агент слушает очередь с метриками и отправляет их на сервер
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/FollowLille/metrics/internal/config"
	"github.com/FollowLille/metrics/internal/crypto"
	"github.com/FollowLille/metrics/internal/logger"
	"github.com/FollowLille/metrics/internal/metrics"
	"github.com/FollowLille/metrics/internal/retry"
	pb "github.com/FollowLille/metrics/proto"
)

type Agent struct {
	ServerAddress      string             // Адрес для прослушивания
	HashKey            string             // Ключ для шифрования
	ServerPort         int64              // Порт для прослушивания
	PollCount          int64              // Количество попыток получения метрик
	RateLimit          int64              // Максимальное количество метрик в секунду
	PollInterval       time.Duration      // Интервал между попытками получения метрик
	ReportSendInterval time.Duration      // Интервал между отправкой метрик
	PublicKey          *rsa.PublicKey     // Публичный ключ для шифрования
	GRPCAddress        string             // Адрес gRPC
	metrics            map[string]float64 // Список метрик
	mutex              sync.Mutex         // Мьютекс для синхронизации доступа к метрикам
	shutdown           chan struct{}      // Канал для остановки агента
	wg                 sync.WaitGroup     // Мьютекс для остановки горутин
}

// NewAgent инициализирует агента
func NewAgent() *Agent {
	return &Agent{
		ServerAddress:      config.Address,
		ServerPort:         config.Port,
		PollInterval:       config.PollInterval,
		ReportSendInterval: config.ReportSendInterval,
		RateLimit:          config.RateLimit,
		metrics:            make(map[string]float64),
		shutdown:           make(chan struct{}),
	}
}

// ChangeIntervalByName изменяет интервал по имени
//
// Параметры:
//   - name - имя интервала
//   - seconds - интервал в секундах
//
// Возвращаемое значение:
//   - error - в случае ошибки
func (a *Agent) ChangeIntervalByName(name string, seconds int64) error {
	if seconds < 1 {
		return fmt.Errorf("incorect inverval value: %d, value must be > 0", seconds)
	}
	newInterval := time.Second * time.Duration(seconds)
	if name == "poll" {
		a.PollInterval = newInterval
	} else if name == "report" {
		a.ReportSendInterval = newInterval
	} else {
		return fmt.Errorf("invalid interval name: %s", name)
	}
	return nil
}

// ChangeAddress изменяет адрес
//
// Параметры:
//   - address - адрес
//
// Возвращаемое значение:
//   - error
func (a *Agent) ChangeAddress(address string) error {
	u, err := url.ParseRequestURI(address)
	if err != nil {
		return err
	}
	if u.Port() != "" {
		return fmt.Errorf("invalid address: %s, port must be empty", address)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("invalid address: %s", address)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("некорректный адрес: %s", address)
	}
	a.ServerAddress = u.Hostname()
	return nil
}

// ChangePort изменяет порт
//
// Параметры:
//   - port - порт
//
// Возвращаемое значение:
//   - error
func (a *Agent) ChangePort(port int64) error {
	if port < 1024 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}
	a.ServerPort = port
	return nil
}

// GetMetrics получает метрики
func (a *Agent) GetMetrics() {
	m := metrics.GetRuntimeMetrics()
	a.mutex.Lock()
	defer a.mutex.Unlock()
	for k, v := range m {
		a.metrics[k] = v
	}
}

// GetGopsutilMetrics получает расширенные метрики
func (a *Agent) GetGopsutilMetrics() {
	m := metrics.GetGopsutilMetrics()
	a.mutex.Lock()
	defer a.mutex.Unlock()
	for k, v := range m {
		a.metrics[k] = v
	}
}

// IncreasePollCount увеличивает счетчик опросов
func (a *Agent) IncreasePollCount() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.PollCount++
}

// Run запускает агента
func (a *Agent) Run() {
	logger.Log.Info("agent running")
	logger.Log.Info("Intervals: ", zap.String("poll", a.PollInterval.String()), zap.String("report", a.ReportSendInterval.String()))
	pollTicker := time.NewTicker(a.PollInterval)
	reportTicker := time.NewTicker(a.ReportSendInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			a.wg.Add(3)
			go func() {
				defer a.wg.Done()
				a.GetMetrics()
			}()
			go func() {
				defer a.wg.Done()
				a.GetGopsutilMetrics()
			}()
			go func() {
				defer a.wg.Done()
				a.IncreasePollCount()
			}()
		case <-reportTicker.C:
			a.wg.Add(1)
			go func() {
				defer a.wg.Done()
				a.ParallelSendMetrics()
			}()
		case <-a.shutdown:
			logger.Log.Info("Shutdown signal received, exiting Run")
			return
		}
	}
}

// ParallelSendMetrics запускает параллельное отправление метрик
func (a *Agent) ParallelSendMetrics() {
	metricsChan := make(chan metrics.Metrics, 30)

	for i := int64(0); i < a.RateLimit; i++ {
		go a.sendByWorker(metricsChan)
	}
	a.mutex.Lock()

	var m metrics.Metrics
	for name, value := range a.metrics {
		m.MType = metrics.Gauge
		m.ID = name
		m.Value = &value
		metricsChan <- m
	}
	m.MType = metrics.Counter
	m.ID = "PollCount"
	delta := a.PollCount
	m.Delta = &delta
	metricsChan <- m

	a.mutex.Unlock()

	close(metricsChan)
}

// sendByWorker отправляет метрики по каналу
// Принимает канал метрик
//
// Параметры:
//   - metricsChan - канал метрик
func (a *Agent) sendByWorker(metricsChan <-chan metrics.Metrics) {
	for m := range metricsChan {
		var err error
		if a.GRPCAddress != "" {
			err = a.sendGRPCMetric(m)
		} else {
			err = a.sendSingleMetric(m)
		}
		if err != nil {
			logger.Log.Error("failed to send single metric", zap.Error(err))
		}
	}
}

// sendSingleMetric отправляет метрику
// Принимает метрику
//
// Параметры:
//   - metric - метрика
//
// Возвращаемое значение:
//   - error
func (a *Agent) sendSingleMetric(metric metrics.Metrics) error {
	jsonMetrics, err := json.Marshal(metric)
	if err != nil {
		logger.Log.Error("failed to marshal metric", zap.String("metric", fmt.Sprintf("%+v", metric)), zap.Error(err))
		return err
	}

	var b bytes.Buffer
	gz, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		logger.Log.Error("failed to create gzip writer", zap.Error(err))
		return err
	}

	if _, err := gz.Write(jsonMetrics); err != nil {
		logger.Log.Error("failed to compress metric", zap.String("metric", fmt.Sprintf("%+v", metric)), zap.Error(err))
		return err
	}

	if err := gz.Flush(); err != nil {
		logger.Log.Error("failed to flush compressed data", zap.Error(err))
		return err
	}
	gz.Close()
	err = retry.Retry(func() error {
		return a.sendRequest(b)
	})
	if err != nil {
		logger.Log.Error("failed to send metric", zap.String("metric", fmt.Sprintf("%+v", metric)), zap.Error(err))
		return err
	}
	return nil
}

// sendRequest отправляет запрос на сервер
// если включен шифрование, то шифруем данные
// если включен хеш, то вычисляем хеш и добавляем его в заголовок
//
// Параметры:
//   - b bytes.Buffer - буфер с данными
//
// Возвращаемое значение:
//   - error
func (a *Agent) sendRequest(b bytes.Buffer) error {
	data := b.Bytes()

	if a.PublicKey != nil {
		encryptedData, err := crypto.Encrypt(a.PublicKey, data)
		if err != nil {
			logger.Log.Error("failed to encrypt data", zap.Error(err))
			return err
		}
		data = encryptedData
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/update", a.ServerAddress, a.ServerPort), bytes.NewReader(data))
	if err != nil {
		logger.Log.Error("failed to create request", zap.Error(err))
		return err
	}
	if a.PublicKey != nil {
		req.Header.Set("Content-Type", "application/octet-stream")
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("X-Real-IP", getLocalIP())

	if a.HashKey != "" {
		hash := crypto.CalculateHash([]byte(a.HashKey), b.Bytes())
		req.Header.Set("HashSHA256", hash)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log.Error("failed to send request", zap.Error(err))
		return retry.ErrorConnection
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 && resp.StatusCode <= 504 {
		logger.Log.Error("received retriable status code", zap.Int("status_code", resp.StatusCode))
		return retry.ErrorServer
	}

	return nil
}

// Shutdown останавливает агента
func (a *Agent) Shutdown() {
	close(a.shutdown)
	logger.Log.Info("Waiting for other workers")
	a.wg.Wait()
	logger.Log.Info("Agent stopped")
}

// getLocalIP возвращает IP локальной машины
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return ""
}

// sendGRPCMetric отправляет метрику
// Принимает метрику
//
// Параметры:
//   - metric - метрика
//
// Возвращаемое значение:
//   - error
func (a *Agent) sendGRPCMetric(metric metrics.Metrics) error {
	pbMetric := &pb.Metric{
		Name:  metric.ID,
		Mtype: metric.MType,
		Delta: metric.Delta,
		Value: metric.Value,
	}

	var encryptedData []byte
	var err error
	if a.PublicKey != nil {
		data, err := proto.Marshal(pbMetric)
		if err != nil {
			logger.Log.Error("failed to marshal metric", zap.String("metric", fmt.Sprintf("%+v", metric)), zap.Error(err))
			return err
		}

		encrypted, err := crypto.Encrypt(a.PublicKey, data)
		if err != nil {
			logger.Log.Error("failed to encrypt data", zap.Error(err))
			return err
		}
		encryptedData = encrypted
	} else {
		encryptedData, err = proto.Marshal(pbMetric)
		if err != nil {
			logger.Log.Error("failed to marshal metric", zap.String("metric", fmt.Sprintf("%+v", metric)), zap.Error(err))
			return err
		}
	}

	conn, err := grpc.NewClient(a.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Error("failed to dial grpc server", zap.Error(err))
		return err
	}
	defer conn.Close()

	client := pb.NewMetricsServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	request := &pb.MetricsRequest{
		Metrics: []*pb.Metric{
			pbMetric,
		},
	}

	if a.HashKey != "" {
		hash := crypto.CalculateHash([]byte(a.HashKey), encryptedData)
		ctx = metadata.AppendToOutgoingContext(ctx, "HashSHA256", hash)
	}

	response, err := client.SendMetrics(ctx, request)
	if err != nil {
		logger.Log.Error("failed to send metric", zap.String("metric", fmt.Sprintf("%+v", metric)), zap.Error(err))
		return err
	}

	logger.Log.Info("sent metric", zap.String("metric", fmt.Sprintf("%+v", metric)))
	logger.Log.Info("response", zap.String("response", fmt.Sprintf("%+v", response)))
	return nil
}
