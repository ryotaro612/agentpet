package config

import (
	"context"
	"log/slog"
	"path/filepath"
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
	// Watch the directory so the watch survives atomic-rename saves (vim,
	// emacs, and most GUI editors write to a temp file then rename it over
	// the original, which would silently drop an inode-level file watch).
	dir := filepath.Dir(w.path)
	if err := fsw.Add(dir); err != nil {
		fsw.Close()
		return err
	}
	base := filepath.Base(w.path)
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
				if filepath.Base(event.Name) != base {
					continue
				}
				if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) && !event.Has(fsnotify.Rename) {
					continue
				}
				cfg, err := LoadConfig(w.path)
				if err != nil {
					w.logger.Warn("config reload failed", "err", err)
					continue
				}
				w.current.Store(&cfg)
				onChange(cfg)
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
