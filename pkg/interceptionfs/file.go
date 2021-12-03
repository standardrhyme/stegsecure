package interceptionfs

import (
	"context"
	"fmt"
	"os"

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

func (f *File) GetNode() (*Node, error) {
	return f.fs.GetNode(f.inum)
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	node, err := f.GetNode()
	if err != nil {
		return err
	}

	*a = node.attr
	mask := ^(0444)
	a.Mode &= os.FileMode(mask)

	fmt.Println(a)
	return nil
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	return &FileHandle{f}, nil
}
