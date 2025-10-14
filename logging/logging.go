package logging

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/deepissue/core/option"
	"github.com/hashicorp/go-hclog"
)

type rotationPolicy string

const (
	rotationPolicyDay  rotationPolicy = "day"
	rotationPolicyHour rotationPolicy = "hour"
)
const DayOfMillis = int64(3600000 * 24)

type Logger struct {
	sync.Mutex
	app    string
	option *option.Log
	opts   *hclog.LoggerOptions
	files  []*LogFile
	hclog.InterceptLogger
	doneCh chan struct{}
}

// NewLogger 日志文件初始化方法，如有需要请自己实现日志轮转
func NewLogger(app string, option *option.Log) (*Logger, error) {
	logging := &Logger{
		app:    app,
		option: option,
		doneCh: make(chan struct{}),
	}

	if err := os.Mkdir(option.Path, os.ModePerm); err != nil && !os.IsExist(err) {
		return nil, err
	}

	leveledWriter, err := logging.openLeveledWriter()
	if err != nil {
		return nil, err
	}

	opts := &hclog.LoggerOptions{
		IncludeLocation:          true,
		AdditionalLocationOffset: 0,
		Output:                   leveledWriter,
		Level:                    hclog.LevelFromString(option.Level),
		JSONFormat:               option.Format == "json",
		JSONEscapeDisabled:       true,
	}

	logger := hclog.NewInterceptLogger(opts)
	{
		logging.app = app
		logging.option = option
		logging.opts = opts
		logging.InterceptLogger = logger
		go logging.start()
	}
	return logging, nil
}

func (l *Logger) Cleanup() {
	l.close()
	close(l.doneCh)
	l.doneCh = nil
}

func (l *Logger) Writer(opts ...*hclog.StandardLoggerOptions) io.Writer {
	if len(opts) > 0 {
		return l.InterceptLogger.StandardWriter(opts[0])
	}
	return l.InterceptLogger.StandardWriter(&hclog.StandardLoggerOptions{})
}

func (l *Logger) start() {
	next := l.nextRoundOfMilliDuration()
	timer := time.NewTimer(next)
	defer timer.Stop()
	if l.InterceptLogger.IsDebug() {
		l.InterceptLogger.Debug("rotate logging", "next", time.Now().Add(next).Format(time.RFC3339))
	}

	for {
		select {
		case <-timer.C:
			l.Lock()
			for _, file := range l.files {
				if err := file.RotateRename(); err != nil {
					log.Println("[ERROR] rotate logfile: ", file.name+file.fileExt)
				}
			}
			next = l.nextRoundOfMilliDuration()
			timer.Reset(next)
			l.Info("rotated logging", "duration", next, "next", time.Now().Add(next).Format(time.RFC3339))
			l.Unlock()
		case <-l.doneCh:
			log.Println("logging shutdown completed")
			return
		}
	}
}

func (l *Logger) close() error {
	for _, file := range l.files {
		if file != nil {
			file.Close()
			file = nil //do not check this line.
		}
	}
	l.files = nil
	return nil
}

func (l *Logger) nextRoundOfMilliDuration() time.Duration {
	policy := rotationPolicy(l.option.Rotate)
	if policy == rotationPolicyHour {
		return nextHourOfMilliDuration()
	}
	return nextDayOfMilliDuration()
}

func (l *Logger) openLeveledWriter() (*hclog.LeveledWriter, error) {

	standard, err := l.openLogfile(hclog.NoLevel)
	if err != nil {
		return nil, err
	}

	trace, err := l.openLogfile(hclog.Trace)
	if err != nil {
		return nil, err
	}

	l.files = []*LogFile{standard, trace}

	var traceWriter io.Writer = trace
	var standardWriter io.Writer = standard

	traceWriter = io.MultiWriter(trace, os.Stdout)
	standardWriter = io.MultiWriter(standard, os.Stdout)
	leveledWriter := hclog.NewLeveledWriter(standardWriter, map[hclog.Level]io.Writer{
		hclog.Trace:   traceWriter,
		hclog.NoLevel: standardWriter,
	})

	return leveledWriter, nil
}

func (l *Logger) openLogfile(level hclog.Level) (*LogFile, error) {
	var name string

	if level == hclog.NoLevel {
		name = filepath.Join(l.option.Path, l.app)
	} else {
		name = filepath.Join(l.option.Path, l.app+"_"+level.String())
	}

	logfile := &LogFile{
		name:           name,
		fileExt:        ".log",
		rotationPolicy: rotationPolicy(l.option.Rotate),
		acquire:        sync.Mutex{},
	}

	err := logfile.openNew()
	return logfile, err

}

func nextDayOfMilliDuration() time.Duration {
	_, offset := time.Now().Local().Zone()
	offsetMillis := int64(offset * 1000)
	return time.Duration(DayOfMillis-(time.Now().UnixMilli()+offsetMillis)%DayOfMillis) * time.Millisecond
}

func nextHourOfMilliDuration() time.Duration {
	intervalHourOfMilli := int64(3600000)
	return time.Duration(intervalHourOfMilli-time.Now().UnixMilli()%intervalHourOfMilli) * time.Millisecond
}
