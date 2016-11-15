package follower

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	_ = fmt.Print
)

type Line struct {
	bytes []byte
}

func (l *Line) Bytes() []byte {
	return l.bytes
}

func (l *Line) String() string {
	return string(l.bytes)
}

type Config struct {
	Offset int64
	Whence int
	Reopen bool
}

type Follower struct {
	once     sync.Once
	file     *os.File
	filename string
	lines    chan Line
	err      error
	config   Config
	reader   *bufio.Reader
	closeCh  chan struct{}
}

func New(filename string, config Config) (*Follower, error) {
	t := &Follower{
		filename: filename,
		lines:    make(chan Line),
		config:   config,
	}

	err := t.reopen()
	if err != nil {
		return nil, err
	}

	go t.once.Do(t.run)

	return t, nil
}

func (t *Follower) Lines() chan Line {
	return t.lines
}

func (t *Follower) Err() error {
	return t.err
}

func (t *Follower) Close() {
	t.closeCh <- struct{}{}
}

func (t *Follower) run() {
	t.close(t.follow())
}

func (t *Follower) follow() error {
	_, err := t.file.Seek(t.config.Offset, t.config.Whence)
	if err != nil {
		return err
	}

	var (
		eventChan = make(chan fsnotify.Event)
		errChan   = make(chan error)
	)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	defer watcher.Close()
	go watchFileEvents(watcher, eventChan, errChan)

	watcher.Add(t.filename)

	for {
		for {
			s, err := t.reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				return err
			}

			// if we encounter EOF before a line delimiter,
			// ReadBytes() will return the remaining bytes,
			// so push them back onto the buffer, rewind
			// our seek position, and wait for further file changes.
			// we also have to save our dangling byte count in the event
			// that we want to re-open the file and seek to the end
			if err == io.EOF {
				l := len(s)

				_, err := t.file.Seek(-int64(l), io.SeekCurrent)
				if err != nil {
					return err
				}

				t.reader.Reset(t.file)

				break
			}

			t.sendLine(s)
		}

		// we're now at EOF, so wait for changes
		select {
		case evt := <-eventChan:
			switch evt.Op {

			// as soon as something is written, go back and read until EOF
			case fsnotify.Write:
				continue

			// truncated. seek to the end minus any dangling bytes before linebreak
			case fsnotify.Chmod:
				_, err = t.file.Seek(0, io.SeekStart)
				if err != nil {
					return err
				}

				t.reader.Reset(t.file)
				continue

			// if a file is removed or renamed
			// and re-opening is desired, see if it appears
			// again within a 1 second deadline. this should be enough time
			// to see the file again for log rotation programs with this behavior
			default:
				if !t.config.Reopen {
					return nil
				}

				watcher.Remove(t.filename)
				time.Sleep(1 * time.Second)

				if err := t.reopen(); err != nil {
					return err
				}

				watcher.Add(t.filename)
				continue
			}

		// any errors that come from fsnotify
		case err := <-errChan:
			return err

		// a request to stop
		case <-t.closeCh:
			watcher.Remove(t.filename)
			return nil

		// fall back to 10 second polling if we haven't received any fsevents
		// stat the file, if it's still there, just continue and try to read bytes
		// if not, go through our re-opening routine
		case <-time.After(10 * time.Second):
			_, err := t.file.Stat()
			if err == nil {
				continue
			}

			if !os.IsNotExist(err) {
				return err
			}

			watcher.Remove(t.filename)
			if err := t.reopen(); err != nil {
				return err
			}

			watcher.Add(t.filename)
			continue
		}
	}

	return nil
}

func (t *Follower) reopen() error {
	if t.file != nil {
		t.file.Close()
		t.file = nil
	}

	file, err := os.Open(t.filename)
	if err != nil {
		return err
	}

	t.file = file
	t.reader = bufio.NewReader(t.file)

	return nil
}

func (t *Follower) close(err error) {
	t.err = err

	if t.file != nil {
		t.file.Close()
	}

	close(t.lines)
}

func (t *Follower) sendLine(l []byte) {
	t.lines <- Line{l[:len(l)-1]}
}

func watchFileEvents(watcher *fsnotify.Watcher, eventChan chan fsnotify.Event, errChan chan error) {
	for {
		select {
		case evt, ok := <-watcher.Events:
			if !ok {
				return
			}

			// debounce write events, but send all others
			switch evt.Op {
			case fsnotify.Write:
				select {
				case eventChan <- evt:
				default:
				}

			default:
				eventChan <- evt
			}

		// die on a file watching error
		case err, _ := <-watcher.Errors:
			errChan <- err
			return
		}
	}
}
