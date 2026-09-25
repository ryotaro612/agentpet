package config

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	path   string
	ch     chan<- Config
	logger *slog.Logger
}

func NewConfigWatcher(ctx context.Context, path string, ch chan<- Config, logger *slog.Logger) (Watcher, error) {
	w := Watcher{path: path, ch: ch, logger: logger}
	if err := w.start(ctx); err != nil {
		return Watcher{}, err
	}
	return w, nil
}

func (w *Watcher) start(ctx context.Context) error {
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
		defer close(w.ch)
		defer fsw.Close()
		for {
			select {
			case <-ctx.Done():
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
				select {
				case w.ch <- cfg:
				case <-ctx.Done():
					return
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
