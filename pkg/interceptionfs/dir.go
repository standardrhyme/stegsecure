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
	children map[string]Node
}

func (d *Dir) FS() *FS { return d.fs }
func (d *Dir) Inum() Inum { return d.inum }
func (d *Dir) Name() string { return d.name }
func (d *Dir) SetName(newName string) { d.name = newName }
func (d *Dir) Parent() *Dir { return d.parent }
func (d *Dir) SetParent(newDir *Dir) { d.parent = newDir }

// GetNode gets the NodeAttr for the current Dir.
func (d *Dir) GetNode() (*NodeAttr, error) {
	return d.fs.GetNode(d.inum)
}

// Attr returns the attributes for the current Dir.
func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	node, err := d.GetNode()
	if err != nil {
		return err
	}

	*a = node.attr
	return nil
}

// LookupNode finds a child Node by name.
func (d *Dir) LookupNode(name string) (Node, *NodeAttr, error) {
	fNode, ok := d.children[name]
	if !ok {
		return nil, nil, syscall.ENOENT
	}

	inum := fNode.Inum()

	node, err := d.fs.GetNode(inum)
	if err != nil {
		return nil, nil, err
	}

	return fNode, node, syscall.ENOENT
}

// Lookup finds a child Node by name, setting additional details.
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

// ReadDirAll lists all the entries in a directory.
func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	node, err := d.GetNode()
	if err != nil {
		return nil, err
	}
	node.UpdateTimes(UATime)

	entries := make([]fuse.Dirent, 0, len(d.children))

	for name, fNode := range d.children {
		inum := fNode.Inum()
		details, err := fNode.GetNode()
		if err != nil {
			return nil, err
		}

		var typ fuse.DirentType
		if details.attr.Mode.IsDir() {
			typ = fuse.DT_Dir
		} else if (details.attr.Mode & os.ModeSymlink) != 0 {
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

// Mkdir creates a new child directory under the current.
func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
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
		children: make(map[string]Node),
	}

	newNode := &NodeAttr{
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

// Create creates a new child file under the current directory.
func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
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

	newNode := &NodeAttr{
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

// Rename renames and/or moves child nodes.
func (d *Dir) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	node, ok := d.children[req.OldName]
	if !ok {
		return fmt.Errorf("Node does not exist.")
	}

	node.SetName(req.NewName)

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

// Remove deletes child nodes.
func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	_, node, err := d.LookupNode(req.Name)
	if err != nil {
		return err
	}

	delete(d.children, req.Name)
	delete(d.fs.nodes, node.GetInum())

	return nil
}
