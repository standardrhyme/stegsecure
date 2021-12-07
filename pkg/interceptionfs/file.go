package interceptionfs

import (
	"context"
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
