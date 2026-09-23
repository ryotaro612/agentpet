package config

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	path    string
	current atomic.Pointer[Config]
	logger  *slog.Logger
	ctx     context.Context
}

func NewConfigWatcher(ctx context.Context, path string, cfg Config, logger *slog.Logger) *Watcher {
	w := &Watcher{path: path, logger: logger, ctx: ctx}
	w.current.Store(&cfg)
	return w
}

func (w *Watcher) Get() Config {
	return *w.current.Load()
}

func (w *Watcher) Watch(onChange func(Config)) error {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := fsw.Add(w.path); err != nil {
		fsw.Close()
		return err
	}
	go func() {
		defer fsw.Close()
		for {
			select {
			case <-w.ctx.Done():
				return
			case event, ok := <-fsw.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					cfg, err := LoadConfig(w.path)
					if err != nil {
						w.logger.Warn("config reload failed", "err", err)
						continue
					}
					w.current.Store(&cfg)
					onChange(cfg)
				}
			case err, ok := <-fsw.Errors:
				if !ok {
					return
				}
				w.logger.Warn("config watcher error", "err", err)
			}
		}
	}()
	return nil
}
