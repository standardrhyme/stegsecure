package interceptionfs

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type Dir struct {
	fs   *FS
	inum Inum

	name     string
	parent   *Dir
	children map[string]fs.Node
}

func (d *Dir) GetNode() (*Node, error) {
	return d.fs.GetNode(d.inum)
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	fmt.Println("getattr", d.inum)
	node, err := d.GetNode()
	if err != nil {
		return err
	}

	*a = node.attr
	return nil
}

func (d *Dir) LookupNode(name string) (fs.Node, *Node, error) {
	for nodeName, fNode := range d.children {
		if name == nodeName {
			var inum Inum
			switch v := fNode.(type) {
			case *Dir:
				inum = v.inum
			case *File:
				inum = v.inum
			default:
				return nil, nil, fmt.Errorf("Unexpected child type.")
			}

			node, err := d.fs.GetNode(inum)
			if err != nil {
				return nil, nil, err
			}

			return fNode, node, nil
		}
	}

	return nil, nil, syscall.ENOENT
}

func (d *Dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	fNode, node, err := d.LookupNode(req.Name)
	if err != nil {
		return nil, err
	}

	resp.Node = fuse.NodeID(node.attr.Inode)
	resp.Attr = node.attr
	resp.EntryValid = node.attr.Valid

	return fNode, nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	node, err := d.GetNode()
	if err != nil {
		return nil, err
	}
	node.UpdateTimes(UATime)

	entries := make([]fuse.Dirent, 0, len(d.children))

	for name, fNode := range d.children {
		var inum Inum
		switch v := fNode.(type) {
		case *Dir:
			inum = v.inum
		case *File:
			inum = v.inum
		default:
			return nil, fmt.Errorf("Unexpected child type.")
		}

		node, err := d.fs.GetNode(inum)
		if err != nil {
			return nil, err
		}

		var typ fuse.DirentType
		if node.attr.Mode.IsDir() {
			typ = fuse.DT_Dir
		} else if (node.attr.Mode & os.ModeSymlink) != 0 {
			typ = fuse.DT_Link
		} else {
			typ = fuse.DT_File
		}

		entries = append(entries, fuse.Dirent{
			Inode: uint64(inum),
			Type:  typ,
			Name:  name,
		})
	}

	return entries, nil
}

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	fmt.Println("MKDIR")
	node, err := d.GetNode()
	if err != nil {
		return nil, err
	}
	node.UpdateTimes(UATime | UMTime)

	inum := d.fs.nextInode.Increment()
	if _, ok := d.fs.nodes[inum]; ok {
		return nil, fmt.Errorf("Out of inodes.")
	}

	newDir := &Dir{
		fs:       d.fs,
		inum:     inum,
		name:     req.Name,
		parent:   d,
		children: make(map[string]fs.Node),
	}

	newNode := &Node{
		fs: d.fs,
		attr: fuse.Attr{
			Mode: req.Mode | os.ModeDir,
			Uid:  uint32(os.Getuid()),
			Gid:  uint32(os.Getgid()),
		},
	}

	newNode.InitAttr(inum)

	d.fs.nodes[inum] = newNode
	d.children[req.Name] = newDir

	fmt.Println(newDir)

	return newDir, nil
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	fmt.Println("create")
	node, err := d.GetNode()
	if err != nil {
		return nil, nil, err
	}
	node.UpdateTimes(UATime | UMTime)

	inum := d.fs.nextInode.Increment()
	if _, ok := d.fs.nodes[inum]; ok {
		return nil, nil, fmt.Errorf("Out of inodes.")
	}

	newFile := &File{
		fs:   d.fs,
		inum: inum,

		name:   req.Name,
		parent: d,
		data:   make([]byte, 0),
	}

	newNode := &Node{
		fs: d.fs,
		attr: fuse.Attr{
			Mode: req.Mode,
			Uid:  uint32(os.Getuid()),
			Gid:  uint32(os.Getgid()),
		},
	}

	newNode.InitAttr(inum)

	d.fs.nodes[inum] = newNode
	d.children[req.Name] = newFile

	return newFile, &FileHandle{newFile}, nil
}

func (d *Dir) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	node, ok := d.children[req.OldName]
	if !ok {
		return fmt.Errorf("Node does not exist.")
	}

	switch n := node.(type) {
	case *Dir:
		n.name = req.NewName
	case *File:
		n.name = req.NewName
	default:
		return fmt.Errorf("Invalid node type.")
	}

	newParentInum := Inum(req.NewDir)

	if newParentInum == d.inum {
		d.children[req.NewName] = d.children[req.OldName]
		delete(d.children, req.OldName)
	} else {
		newParent, ok := newDir.(*Dir)
		if !ok {
			return fmt.Errorf("New parent is not a directory.")
		}

		newParent.children[req.NewName] = d.children[req.OldName]
		delete(d.children, req.OldName)
	}

	return nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	node, ok := d.children[req.Name]
	if !ok {
		return fmt.Errorf("Node does not exist.")
	}

	var inum Inum
	switch n := node.(type) {
	case *Dir:
		inum = n.inum
	case *File:
		inum = n.inum
	default:
		return fmt.Errorf("Invalid node type.")
	}

	delete(d.children, req.Name)
	delete(d.fs.nodes, inum)

	return nil
}
