package interceptionfs

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"bazil.org/fuse"
)

type FileHandle struct {
	*File
	passthroughHandle *os.File
}

// InternalRead allows internal access to the bytes in the file, not guarded by
// passthrough or cleaned.
func (fh *FileHandle) InternalRead(req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	if fh.passthrough {
		data := make([]byte, req.Size)
		bytesRead, err := fh.passthroughHandle.ReadAt(data, req.Offset)
		if err != nil {
			return err
		}

		resp.Data = data[:bytesRead]
		return nil
	}

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
	if fh.cleaned || fh.passthrough {
		return fh.InternalRead(req, resp)
	}
	return syscall.EPERM
}

// InternalReadAll reads the entire file.
func (fh *FileHandle) InternalReadAll() ([]byte, error) {
	if fh.passthrough {
		return os.ReadFile(fh.GetRealPath())
	}
	return fh.data, nil
}

// ReadAll reads the entire file.
func (fh *FileHandle) ReadAll(ctx context.Context) ([]byte, error) {
	if fh.fs.RemoveIfNotExist(fh) {
		return nil, syscall.ENOENT
	}

	if fh.cleaned || fh.passthrough {
		return fh.InternalReadAll()
	}
	return nil, syscall.EPERM
}

func (fh *FileHandle) InternalOverwrite(data []byte) {
	if fh.passthrough {
		return
	}

	fh.data = data
}

// Write modifies the contents of the file.
func (fh *FileHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	if fh.fs.RemoveIfNotExist(fh) {
		return syscall.ENOENT
	}

	node, err := fh.GetNode()
	if err != nil {
		return err
	}

	// If the file was a passthrough file, copy it into the filesystem.
	if fh.passthrough {
		if err := fh.parent.ResolvePassthrough(); err != nil {
			return err
		}

		info, err := os.Stat(fh.GetRealPath())
		if err != nil {
			return err
		}

		oldInum := fh.inum
		inum := fh.fs.nextInum.Increment()
		if _, ok := fh.fs.nodes[inum]; ok {
			return fmt.Errorf("Out of inodes.")
		}

		data, err := fh.InternalReadAll()
		if err != nil {
			return err
		}

		fh.inum = inum
		fh.data = data
		fh.cleaned = false
		fh.passthrough = false

		fh.passthroughHandle.Close()
		fh.passthroughHandle = nil

		newNode := *node
		node = &newNode
		node.InitAttr(inum)
		node.attr.Size = uint64(info.Size())
		node.attr.Mode = info.Mode()
		node.attr.Mtime = info.ModTime()

		fh.fs.nodes[inum] = node
		delete(fh.fs.passNodes, oldInum)
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

	go fh.fs.GetNotifier(fh.inum)(fh)

	return nil
}

func (fh *FileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	if fh.fs.RemoveIfNotExist(fh) {
		return syscall.ENOENT
	}

	if !fh.passthrough {
		return nil
	}

	return fh.passthroughHandle.Close()
}
