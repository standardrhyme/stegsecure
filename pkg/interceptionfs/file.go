package interceptionfs

import (
	"context"
	// "os"

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

func (f *File) FS() *FS { return f.fs }
func (f *File) Inum() Inum { return f.inum }
func (f *File) Name() string { return f.name }
func (f *File) SetName(newName string) { f.name = newName }
func (f *File) Parent() *Dir { return f.parent }
func (f *File) SetParent(newDir *Dir) { f.parent = newDir }

// GetNode gets the NodeAttr for the current File.
func (f *File) GetNode() (*NodeAttr, error) {
	return f.fs.GetNode(f.inum)
}

// Attr returns the attributes for the current File.
func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	node, err := f.GetNode()
	if err != nil {
		return err
	}

	// Remove all read permissions.
	*a = node.attr
	// mask := ^(0444)
	// a.Mode &= os.FileMode(mask)

	return nil
}

// Open opens a handle to a File, for reading or writing.
func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	return &FileHandle{f}, nil
}
