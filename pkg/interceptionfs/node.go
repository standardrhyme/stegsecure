package interceptionfs

import (
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type Inum uint64

func (i *Inum) Increment() Inum {
	temp := *i
	*i = temp + 1
	return temp
}

type UpdateTime uint8

const (
	UATime UpdateTime = 1 << (2 - iota)
	UMTime
	UCTime

	UAllTime = UATime | UMTime | UCTime
)

// Node is an interface defining common fields between Dir and File
type Node interface {
	fs.Node

	FS() *FS
	Inum() Inum
	Name() string
	SetName(newName string)
	Parent() *Dir
	SetParent(newDir *Dir)
	GetNode() (*NodeAttr, error)
}

// NodeAttr stores the attributes of a node.
type NodeAttr struct {
	fs *FS
	attr fuse.Attr
}

func (n *NodeAttr) GetInum() Inum {
	return Inum(n.attr.Inode)
}

func (n *NodeAttr) InitAttr(inode Inum) {
	n.attr.Valid = time.Minute
	n.attr.Inode = uint64(inode)
	n.attr.Nlink = 1

	n.UpdateTimes(UAllTime)
}

/*
func (n *Node) Symlink() (*Symlink, error) {
	if n.attr.Mode.IsDir() || (n.attr.Mode & os.ModeSymlink) == 0 {
		return nil, fmt.Errorf("Node is not a symlink.")
	}

	symlink, ok := n.Node.(*Symlink)
	if !ok {
		return nil, fmt.Errorf("Node was recorded to be a symlink, but is not.")
	}

	return symlink, nil
}
*/

func (n *NodeAttr) UpdateTimes(updates UpdateTime) {
	now := time.Now()

	if (updates & UATime) != 0 {
		n.attr.Atime = now
	}

	if (updates & UMTime) != 0 {
		n.attr.Mtime = now
	}

	if (updates & UCTime) != 0 {
		n.attr.Ctime = now
	}
}
