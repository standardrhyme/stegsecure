package interceptionfs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type FS struct {
	Debug func(msg interface{})
	conn *fuse.Conn
	mountpoint string

	rootInum  Inum
	root      fs.Node
	nodes     map[Inum]*NodeAttr
	nextInode Inum
	notifier  func(Node)
}

// Init sets up the filesystem by creating the variables needed, as well as the
// root directory.
func Init(notifier func(Node)) (*FS, error) {
	f := &FS{
		nodes:     make(map[Inum]*NodeAttr),
		nextInode: 1,
		notifier:  DebouncedNotifier(notifier, time.Second),
	}

	f.Debug = func(msg interface{}){}

	// Give the root directory the first inum.
	rootNum := f.nextInode.Increment()

	f.rootInum = rootNum

	f.nodes[rootNum] = &NodeAttr{
		fs: f,
		attr: fuse.Attr{
			Mode: os.ModeDir | 0755,
			Uid:  uint32(os.Getuid()),
			Gid:  uint32(os.Getgid()),
		},
	}

	f.nodes[rootNum].InitAttr(rootNum)

	f.root = &Dir{
		fs:       f,
		inum:     rootNum,
		name:     "",
		parent:   nil,
		children: make(map[string]Node),
	}

	return f, nil
}

// Mount creates a fuse connection at the destination.
func (f *FS) Mount(mountpoint string) error {
	options := make([]fuse.MountOption, 1)
	options[0] = fuse.AllowNonEmptyMount()
	if os.Geteuid() == 0 {
		options = append(options, fuse.AllowOther())
	}

	var err error
	f.mountpoint, err = filepath.Abs(mountpoint)
	if err != nil {
		return err
	}

	c, err := fuse.Mount(mountpoint, options...)
	if err != nil {
		return err
	}

	f.conn = c
	return nil
}

// Serve serves the filesystem to a fuse connection.
func (f *FS) Serve(res chan error) error {
	if f.conn == nil {
		return fmt.Errorf("Connection must be opened first, using Mount.")
	}

	go func() {
		server := fs.New(f.conn, &fs.Config{
			Debug: f.Debug,
		})
		res <- server.Serve(f)
	}()

	return nil
}

// Close closes a fuse connection.
func (f *FS) Close() error {
	if f.conn == nil {
		return fmt.Errorf("Connection is not open.")
	}
	return f.conn.Close()
}

// GetNode returns the node from a given inum.
func (f *FS) GetNode(inum Inum) (*NodeAttr, error) {
	node, ok := f.nodes[inum]
	if !ok {
		return nil, fmt.Errorf("Inum not found in map")
	}

	return node, nil
}

// Root returns the root directory.
func (f *FS) Root() (fs.Node, error) {
	return f.root, nil
}
