package watcher

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/arnarg/mads/pkg/entities"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

const (
	TypeApply  = "apply"
	TypeDelete = "delete"
)

type PodFileEvent struct {
	Type string
	Name string
	Pod  *entities.Pod
}

type FileWatcher struct {
	path string
	ch   chan *PodFileEvent
	pods map[string]*entities.Pod
}

func NewFileWatcher(p string) *FileWatcher {
	return &FileWatcher{
		path: p,
		ch:   make(chan *PodFileEvent, 100),
		pods: map[string]*entities.Pod{},
	}
}

func (w *FileWatcher) Run(ctx context.Context) error {
	// Before watching we want to parse all files in the directory
	files, err := os.ReadDir(w.path)
	if err != nil {
		return err
	}

	// Read all files in directory
	for _, f := range files {
		// Skip directories
		if f.IsDir() {
			continue
		}

		err := w.parseFile(fmt.Sprintf("%s/%s", w.path, f.Name()))
		if err != nil {
			return err
		}
	}

	// Create a fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Watch watch-dir
	err = watcher.Add(w.path)
	if err != nil {
		return fmt.Errorf("could not watch directory '%s': %s", w.path, err)
	}

	// Wait for events
	for {
		select {
		// File event
		case ev, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher channel was closed")
			}

			switch {
			// File created or updated
			case ev.Op == fsnotify.Create || ev.Op == fsnotify.Write:
				err := w.parseFile(ev.Name)
				if err != nil {
					return err
				}

			// File renamed
			case ev.Op == fsnotify.Rename:
				// When files are renamed two events are fired, first a rename event with old file name
				// and then a create event with the new file name.
				// I don't want to delete the pod on renames and then create it again so I just remove
				// the pod from the pods map and then run apply again when I receive the create event
				// above, which should be a no-op.
				delete(w.pods, ev.Name)

			// File removed
			case ev.Op == fsnotify.Remove:
				// Get old pod from saved pods map
				pod, ok := w.pods[ev.Name]
				if !ok {
					// We ignore it
					continue
				}

				// Delete the pod from the pods map
				delete(w.pods, ev.Name)

				// Send a delete event to channel
				w.ch <- &PodFileEvent{
					Type: TypeDelete,
					Name: pod.Name,
				}
			}

		// Watcher error
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher channel was closed")
			}

			fmt.Println(err)

		// Context cancelled
		case <-ctx.Done():
			return nil
		}
	}
}

func (w *FileWatcher) parseFile(p string) error {
	// Read file
	buf, err := ioutil.ReadFile(p)
	if err != nil {
		return fmt.Errorf("could not read file '%s': %s", p, err)
	}

	// Parse file
	pod := &entities.Pod{}
	err = yaml.Unmarshal(buf, pod)
	if err != nil {
		return fmt.Errorf("could not parse file '%s': %s", p, err)
	}

	// Save pod in map.
	// This is necessary so we can get the name of the pod when a file is deleted.
	w.pods[p] = pod

	// Send event to channel
	w.ch <- &PodFileEvent{
		Type: TypeApply,
		Name: pod.Name,
		Pod:  pod,
	}

	return nil
}

func (w *FileWatcher) PodFileEvents() <-chan *PodFileEvent {
	return w.ch
}
