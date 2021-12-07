package interceptionfs

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type Dir struct {
	fs   *FS
	inum Inum

	name        string
	parent      *Dir
	children    map[string]Node
	passthrough bool
}

func (d *Dir) FS() *FS                { return d.fs }
func (d *Dir) Inum() Inum             { return d.inum }
func (d *Dir) Name() string           { return d.name }
func (d *Dir) SetName(newName string) { d.name = newName }
func (d *Dir) Parent() *Dir           { return d.parent }
func (d *Dir) SetParent(newDir *Dir)  { d.parent = newDir }
func (d *Dir) Passthrough() bool      { return d.passthrough }
func (d *Dir) GetRelPath() string {
	if d.inum == d.fs.rootInum {
		return ""
	}
	return d.parent.GetRelPath() + "/" + d.name
}
func (d *Dir) GetRealPath() string { return d.fs.GetRealPath(d.GetRelPath()) }

// GetNode gets the NodeAttr for the current Dir.
func (d *Dir) GetNode() (*NodeAttr, error) {
	return d.fs.GetNode(d.inum)
}

// Attr returns the attributes for the current Dir.
func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	if d.fs.RemoveIfNotExist(d) {
		return syscall.ENOENT
	}

	node, err := d.GetNode()
	if err != nil {
		return err
	}

	*a = node.attr

	if d.passthrough {
		info, err := os.Stat(d.GetRealPath())
		if err != nil {
			return err
		}

		a.Size = uint64(info.Size())
		a.Mode = info.Mode()
		a.Mtime = info.ModTime()
	}

	return nil
}

// Lookup finds a child Node by name, setting additional details.
func (d *Dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	if d.fs.RemoveIfNotExist(d) {
		return nil, syscall.ENOENT
	}

	if err := d.UpdatePassthroughChildren(); err != nil {
		return nil, err
	}

	fNode, ok := d.children[req.Name]
	if !ok {
		return nil, syscall.ENOENT
	}

	node, err := fNode.GetNode()
	if err != nil {
		return nil, err
	}

	resp.Node = fuse.NodeID(fNode.Inum())
	resp.Attr = node.attr
	resp.EntryValid = node.attr.Valid

	return fNode, nil
}

// ReadDirAll lists all the entries in a directory.
func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	if d.fs.RemoveIfNotExist(d) {
		return nil, syscall.ENOENT
	}

	if err := d.UpdatePassthroughChildren(); err != nil {
		return nil, err
	}

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
	if d.fs.RemoveIfNotExist(d) {
		return nil, syscall.ENOENT
	}

	if err := d.ResolvePassthrough(); err != nil {
		return nil, err
	}

	node, err := d.GetNode()
	if err != nil {
		return nil, err
	}
	node.UpdateTimes(UATime | UMTime)

	inum := d.fs.nextInum.Increment()
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

	return newDir, nil
}

// Create creates a new child file under the current directory.
func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	if d.fs.RemoveIfNotExist(d) {
		return nil, nil, syscall.ENOENT
	}

	if err := d.ResolvePassthrough(); err != nil {
		return nil, nil, err
	}

	node, err := d.GetNode()
	if err != nil {
		return nil, nil, err
	}
	node.UpdateTimes(UATime | UMTime)

	inum := d.fs.nextInum.Increment()
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

	return newFile, &FileHandle{newFile, nil}, nil
}

// Rename renames and/or moves child nodes.
func (d *Dir) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	if d.fs.RemoveIfNotExist(d) {
		return syscall.ENOENT
	}

	node, ok := d.children[req.OldName]
	if !ok {
		return syscall.ENOENT
	}

	if d.fs.RemoveIfNotExist(node) {
		return syscall.ENOENT
	}

	oldPath := node.GetRealPath()

	node.SetName(req.NewName)

	newParentInum := Inum(req.NewDir)

	if newParentInum == d.inum {
		d.children[req.NewName] = d.children[req.OldName]
	} else {
		newParent, ok := newDir.(*Dir)
		if !ok {
			return syscall.ENOENT
		}

		newParent.children[req.NewName] = d.children[req.OldName]
	}

	if node.Passthrough() {
		newPath := node.GetRealPath()
		if err := os.Rename(oldPath, newPath); err != nil {
			return err
		}
	}

	delete(d.children, req.OldName)

	return nil
}

// Remove deletes child nodes.
func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	if d.fs.RemoveIfNotExist(d) {
		return syscall.ENOENT
	}

	node, ok := d.children[req.Name]
	if !ok {
		return syscall.ENOENT
	}

	if d.fs.RemoveIfNotExist(node) {
		return syscall.ENOENT
	}

	var m map[Inum]*NodeAttr

	if node.Passthrough() {
		if err := os.Remove(node.GetRealPath()); err != nil {
			return err
		}
		m = d.fs.passNodes
	} else {
		m = d.fs.nodes
	}

	delete(d.children, req.Name)
	delete(m, node.Inum())

	return nil
}

func (d *Dir) ResolvePassthrough() error {
	if !d.passthrough || d.inum == d.fs.rootInum {
		return nil
	}

	node, err := d.GetNode()
	if err != nil {
		return err
	}

	oldInum := d.inum
	inum := d.fs.nextInum.Increment()
	if _, ok := d.fs.nodes[inum]; ok {
		return fmt.Errorf("Out of inodes.")
	}

	d.inum = inum
	d.passthrough = false

	newNode := *node
	newNode.InitAttr(inum)

	d.fs.nodes[inum] = &newNode
	delete(d.fs.passNodes, oldInum)

	return d.parent.ResolvePassthrough()
}

func (d *Dir) UpdatePassthroughChildren() error {
	oldPassthrough := make(map[string]Node)

	for name, fNode := range d.children {
		if fNode.Passthrough() {
			oldPassthrough[name] = fNode
		}
	}

	realPath := d.GetRealPath()

	realFiles, err := ioutil.ReadDir(realPath)
	if err != nil {
		return err
	}

	for _, fileInfo := range realFiles {
		var inum Inum
		name := fileInfo.Name()

		if node, ok := d.children[name]; ok && !node.Passthrough() {
			continue
		}

		node, ok := oldPassthrough[name]
		if ok {
			inum = node.Inum()
			delete(oldPassthrough, name)
		} else {
			var newNode Node
			inum = d.fs.nextPassInum.Decrement()
			if _, ok := d.fs.passNodes[inum]; ok {
				break
			}

			if fileInfo.IsDir() {
				newNode = &Dir{
					fs:          d.fs,
					inum:        inum,
					name:        name,
					parent:      d,
					children:    make(map[string]Node),
					passthrough: true,
				}
			} else {
				newNode = &File{
					fs:          d.fs,
					inum:        inum,
					name:        name,
					parent:      d,
					passthrough: true,
				}
			}

			newNodeAttr := &NodeAttr{
				fs: d.fs,
				attr: fuse.Attr{
					Size: uint64(fileInfo.Size()),
					Mode: fileInfo.Mode(),
				},
			}

			newNodeAttr.InitAttr(inum)
			newNodeAttr.attr.Mtime = fileInfo.ModTime()

			stat, ok := fileInfo.Sys().(*syscall.Stat_t)
			if ok {
				newNodeAttr.attr.Uid = stat.Uid
				newNodeAttr.attr.Gid = stat.Gid
				newNodeAttr.attr.Atime = time.Unix(stat.Atim.Sec, stat.Atim.Nsec)
				newNodeAttr.attr.Ctime = time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)
			}

			d.fs.passNodes[inum] = newNodeAttr
			d.children[name] = newNode
		}
	}

	// Delete any passthrough children that no longer exist.
	for name, node := range oldPassthrough {
		delete(d.children, name)
		delete(d.fs.passNodes, node.Inum())
	}

	return nil
}
