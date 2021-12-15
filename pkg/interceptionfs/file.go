package interceptionfs

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type File struct {
	fs   *FS
	inum Inum

	name   string
	parent *Dir
	data   []byte

	cleaned     bool
	passthrough bool
}

func (f *File) FS() *FS                { return f.fs }
func (f *File) Inum() Inum             { return f.inum }
func (f *File) Name() string           { return f.name }
func (f *File) SetName(newName string) { f.name = newName }
func (f *File) Parent() *Dir           { return f.parent }
func (f *File) SetParent(newDir *Dir)  { f.parent = newDir }
func (f *File) Passthrough() bool      { return f.passthrough }
func (f *File) GetRelPath() string {
	return f.parent.GetRelPath() + "/" + f.name
}
func (f *File) GetRealPath() string { return f.fs.GetRealPath(f.GetRelPath()) }

// GetNode gets the NodeAttr for the current File.
func (f *File) GetNode() (*NodeAttr, error) {
	return f.fs.GetNode(f.inum)
}

func (f *File) SetCleaned(cleaned bool) {
	f.cleaned = cleaned
}

// Attr returns the attributes for the current File.
func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	if f.fs.RemoveIfNotExist(f) {
		return syscall.ENOENT
	}

	node, err := f.GetNode()
	if err != nil {
		return err
	}

	*a = node.attr

	if f.passthrough {
		info, err := os.Stat(f.GetRealPath())
		if err != nil {
			return err
		}

		a.Size = uint64(info.Size())
		a.Mode = info.Mode()
		a.Mtime = info.ModTime()
	}

	if !f.passthrough && !f.cleaned {
		// Remove all read permissions.
		mask := ^(0444)
		a.Mode &= os.FileMode(mask)
	}

	return nil
}

// Open opens a handle to a File, for reading or writing.
func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	if f.fs.RemoveIfNotExist(f) {
		return nil, syscall.ENOENT
	}

	var file *os.File
	var err error
	if f.passthrough {
		file, err = os.OpenFile(f.GetRealPath(), int(req.Flags), 0)
		if err != nil {
			return nil, err
		}
	}

	return &FileHandle{f, file}, nil
}

func (f *File) Release() error {
	node, err := f.GetNode()
	if err != nil {
		return err
	}

	inum := f.fs.nextPassInum.Decrement()
	if _, ok := f.fs.passNodes[inum]; ok {
		return fmt.Errorf("Out of inodes")
	}

	path := f.GetRealPath()

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	err = os.Chmod(path, node.attr.Mode.Perm())
	if err != nil {
		return err
	}

	err = os.Chown(path, int(node.attr.Uid), int(node.attr.Gid))
	if err != nil {
		return err
	}

	err = os.Chtimes(path, node.attr.Atime, node.attr.Ctime)
	if err != nil {
		return err
	}

	_, err = file.Write(f.data)
	if err != nil {
		return err
	}

	oldInum := f.inum

	f.inum = inum
	f.data = nil

	f.cleaned = false
	f.passthrough = true

	f.fs.passNodes[inum] = node
	delete(f.fs.nodes, oldInum)

	return nil
}
