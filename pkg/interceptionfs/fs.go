package interceptionfs

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type FS struct {
	Debug      func(msg interface{})
	conn       *fuse.Conn
	mountpoint string
	binddir    string

	rootInum  Inum
	root      fs.Node
	nodes     map[Inum]*NodeAttr
	notifiers map[Inum]*func(Node)
	nextInum Inum
	notifier  func(Node)

	nextPassInum Inum
	passNodes    map[Inum]*NodeAttr
}

// Init sets up the filesystem by creating the variables needed, as well as the
// root directory.
func Init(notifier func(Node)) (*FS, error) {
	f := &FS{
		nodes:     make(map[Inum]*NodeAttr),
		notifiers: make(map[Inum]*func(Node)),
		nextInum: 1,
		notifier:  notifier,

		nextPassInum: math.MaxUint64,
		passNodes:    make(map[Inum]*NodeAttr),
	}

	f.Debug = func(msg interface{}) {}

	// Give the root directory the first inum.
	rootNum := f.nextInum.Increment()

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
	var success bool

	options := []fuse.MountOption{
		fuse.AllowNonEmptyMount(),
		fuse.AllowOther(),
	}

	c, err := fuse.Mount(mountpoint, options...)
	if err != nil {
		return err
	}

	f.conn = c
	defer func() {
		if !success {
			f.conn.Close()
		}
	}()

	parentDir, err := filepath.Abs(filepath.Join(mountpoint, ".."))
	if err != nil {
		return err
	}

	f.binddir, err = os.MkdirTemp("", "*")
	if err != nil {
		return err
	}

	cmd := exec.Command("mount", "--bind", parentDir, f.binddir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	f.mountpoint = filepath.Join(f.binddir, filepath.Base(mountpoint))
	fmt.Println(parentDir, f.mountpoint)

	success = true

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

	umount := exec.Command("umount", f.binddir)
	umount.Stdout = os.Stdout
	umount.Stderr = os.Stderr

	delBindDir := exec.Command("rm", "-rf", f.binddir)
	delBindDir.Stdout = os.Stdout
	delBindDir.Stderr = os.Stderr

	errors := make([]error, 0)

	if err := f.conn.Close(); err != nil {
		errors = append(errors, err)
	}

	if f.binddir != "" {
		if err := umount.Run(); err != nil {
			errors = append(errors, err)
		} else if err := delBindDir.Run(); err != nil {
			errors = append(errors, err)
		}
	}

	switch len(errors) {
	case 0:
		return nil
	case 1:
		return errors[0]
	}
	return fmt.Errorf("Errors: %w, %v", errors[0], errors[1:])
}

// GetNode returns the node from a given inum.
func (f *FS) GetNode(inum Inum) (*NodeAttr, error) {
	if node, ok := f.nodes[inum]; ok {
		return node, nil
	}

	if node, ok := f.passNodes[inum]; ok {
		return node, nil
	}

	return nil, syscall.ENOENT
}

// Gets the path to the real path, given a path relative to the root of the mount.
func (f *FS) GetRealPath(relPath string) string {
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(f.mountpoint, "/"), strings.TrimPrefix(relPath, "/"))
}

// GetNotifier creates and stores a DebouncedNotifier specific to the inum.
func (f *FS) GetNotifier(inum Inum) func(Node) {
	if n, ok := f.notifiers[inum]; ok {
		return *n
	}

	n := DebouncedNotifier(f.notifier, time.Second)
	f.notifiers[inum] = &n
	return n
}

func (f *FS) Exists (relPath string) bool {
	_, err := os.Stat(relPath)
	return !os.IsNotExist(err)
}

func (f *FS) RemoveIfNotExist(n Node) bool {
	if n.Inum() == f.rootInum { return false }
	if !n.Passthrough() { return false }
	if f.Exists(n.GetRealPath()) { return false }

	delete(n.Parent().children, n.Name())
	delete(f.passNodes, n.Inum())

	return true
}

// Root returns the root directory.
func (f *FS) Root() (fs.Node, error) {
	return f.root, nil
}
