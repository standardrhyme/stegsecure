package interceptionfs

import (
	"context"
	"fmt"

	"bazil.org/fuse"
)

type FileHandle struct {
	*File
}

// InternalRead allows internal access to the bytes in the file, not guarded by
// passthrough or cleaned.
func (fh *FileHandle) InternalRead(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	node, err := fh.GetNode()
	if err != nil {
		return err
	}

	node.UpdateTimes(UATime)

	size := uint64(req.Size)
	if size+uint64(req.Offset) > node.attr.Size {
		size = node.attr.Size - uint64(req.Offset)
	}

	resp.Data = fh.data[req.Offset : uint64(req.Offset)+size]

	return nil
}

// Read allows file system clients to read the file, if it has been cleaned.
func (fh *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	if fh.passthrough {
		// passthrough read
	}
	if fh.cleaned {
		return fh.InternalRead(ctx, req, resp)
	}
	return fmt.Errorf("No permission")
}

// Write modifies the contents of the file.
func (fh *FileHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	node, err := fh.GetNode()
	if err != nil {
		return err
	}

	node.UpdateTimes(UATime | UMTime)
	resp.Size = len(req.Data)

	oldSize := node.attr.Size
	writeEndIndex := Uint64Max(uint64(req.Offset)+uint64(resp.Size), oldSize)
	fileEndIndex := Uint64Max(writeEndIndex, oldSize)

	var dataPost []byte
	dataPre := fh.data[0:req.Offset]

	if fileEndIndex < oldSize {
		dataPost = fh.data[writeEndIndex:fileEndIndex]
	}

	fh.data = append(dataPre, req.Data...)
	fh.data = append(fh.data, dataPost...)
	node.attr.Size = fileEndIndex

	go fh.fs.notifier(fh)

	return nil
}
