package logging

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/orion/elastic"
	es "gopkg.in/olivere/elastic.v5"
)

type LoggingService struct {
	bufferChan chan []byte
	buffer     [][]byte
	config     *elastic.Conf
	logger     *logger.Logger
	timer      *Timer
	closeCh    chan struct{}
	lock       sync.Mutex
	client     *es.Client
}

//NewLoggingService 创建日志组件
func NewLoggingService(client *es.Client, c *elastic.Conf) (r *LoggingService, err error) {
	r = &LoggingService{
		client:     client,
		config:     c,
		logger:     logger.GetSession("orion.logging", logger.CreateSession()),
		bufferChan: make(chan []byte, 100000),
		buffer:     make([][]byte, 0, 1000),
		closeCh:    make(chan struct{}),
	}
	if r.timer, err = NewTimer(c.Cron); err != nil {
		return nil, err
	}
	r.timer.Start()
	go r.loopWrite()
	return r, nil
}

//Save 保存日志
func (l *LoggingService) Save(data string) error {
	var buff bytes.Buffer
	if err := json.Compact(&buff, []byte(data)); err != nil {
		return err
	}
	l.bufferChan <- buff.Bytes()
	return nil
}
func (l *LoggingService) loopWrite() {
	notify := l.timer.Subscribe()
	for {
		select {
		case <-l.closeCh:
			return
		case v := <-l.bufferChan:
			if len(v) <= 2 {
				continue
			}
			l.lock.Lock()
			if v[0] == '[' {
				l.buffer = append(l.buffer, bytes.Split(v[1:len(v)-1], []byte(","))...)
			} else {
				l.buffer = append(l.buffer, v)
			}
			l.lock.Unlock()

		case <-l.closeCh:
			l.lock.Lock()
			if len(l.buffer) <= 0 {
				l.lock.Unlock()
				continue
			}
			go l.Write(l.buffer[0:])
			l.buffer = l.buffer[:0]
			l.lock.Unlock()
			return
		case <-notify:
			l.lock.Lock()
			if len(l.buffer) <= 0 {
				l.lock.Unlock()
				continue
			}

			go l.Write(l.buffer[0:])
			l.buffer = l.buffer[:0]
			l.lock.Unlock()

		}
	}
}
func (l *LoggingService) Write(p [][]byte) (n int, err error) {
	if err := elastic.BenchAddData(l.client, l.config.TypeName, l.config.Index, l.config.WriteTimeout, p); err != nil {
		l.logger.Error(err)
		return 0, err
	}
	l.logger.Debugf("-> write log ", len(p))
	return len(p) - 1, nil
}

//Close 关闭当前日志组件
func (l *LoggingService) Close() error {
	if l.timer != nil {
		l.timer.Close()
	}
	close(l.closeCh)

	return nil
}
