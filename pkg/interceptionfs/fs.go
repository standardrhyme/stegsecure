package interceptionfs

import (
	"fmt"
	"os"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type FS struct {
	Debug func(msg interface{})
	conn *fuse.Conn

	rootInum  Inum
	root      fs.Node
	nodes     map[Inum]*Node
	nextInode Inum
	notifier  func(*Node)
}

func Init(notifier func(*Node)) (*FS, error) {
	f := &FS{
		nodes:     make(map[Inum]*Node),
		nextInode: 1,
		notifier:  DebouncedNotifier(notifier, time.Second),
	}

	f.Debug = func(msg interface{}){}

	rootNum := f.nextInode.Increment()

	f.rootInum = rootNum

	f.nodes[rootNum] = &Node{
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
		children: make(map[string]fs.Node),
	}

	return f, nil
}

func (f *FS) Mount(mountpoint string) error {
	options := make([]fuse.MountOption, 1)
	options[0] = fuse.AllowNonEmptyMount()
	if os.Geteuid() == 0 {
		options = append(options, fuse.AllowOther())
	}

	c, err := fuse.Mount(mountpoint, options...)
	if err != nil {
		return err
	}

	f.conn = c
	return nil
}

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

func (f *FS) Close() error {
	if f.conn == nil {
		return fmt.Errorf("Connection is not open.")
	}
	return f.conn.Close()
}

func (f *FS) GetNode(inum Inum) (*Node, error) {
	node, ok := f.nodes[inum]
	if !ok {
		return nil, fmt.Errorf("Inum not found in map")
	}

	return node, nil
}

func (f *FS) Root() (fs.Node, error) {
	return f.root, nil
}
