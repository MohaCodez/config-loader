package loader

import (
	"fmt"
	"os"
	"time"
	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher  *fsnotify.Watcher
	callback func()
	onError  func(error)
	done     chan struct{}
}

func NewWatcher(paths []string, callback func(), onError func(error)) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		if err := fsw.Add(path); err != nil {
			fsw.Close()
			return nil, err
		}
	}

	w := &Watcher{
		watcher:  fsw,
		callback: callback,
		onError:  onError,
		done:     make(chan struct{}),
	}

	go w.watch()
	return w, nil
}

func (w *Watcher) watch() {
	var timer *time.Timer
	debounce := 300 * time.Millisecond

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(debounce, w.callback)
			}
			
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			if err != nil {
				if w.onError != nil {
					w.onError(err)
				} else {
					fmt.Fprintf(os.Stderr, "watcher error: %v\n", err)
				}
			}
			
		case <-w.done:
			return
		}
	}
}

func (w *Watcher) Close() error {
	close(w.done)
	return w.watcher.Close()
}
