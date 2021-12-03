package interceptionfs

import (
	// "context"
	// "fmt"
	// "os"
	"time"

	"bazil.org/fuse"
	// "bazil.org/fuse/fs"
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

type Node struct {
	// fs.Node
	fs *FS

	// parent Inum

	attr fuse.Attr
}

func (n *Node) GetInum() Inum {
	return Inum(n.attr.Inode)
}

func (n *Node) InitAttr(inode Inum) {
	n.attr.Valid = time.Minute
	n.attr.Inode = uint64(inode)
	n.attr.Nlink = 1

	n.UpdateTimes(UAllTime)
}

// func (n *Node) Dir() (*Dir, error) {
// 	if !n.attr.Mode.IsDir() {
// 		return nil, fmt.Errorf("Node is not a directory.")
// 	}
//
// 	dir, ok := n.Node.(*Dir)
// 	if !ok {
// 		return nil, fmt.Errorf("Node was recorded to be a directory, but is not.")
// 	}
//
// 	return dir, nil
// }
//
// func (n *Node) File() (*File, error) {
// 	if n.attr.Mode.IsDir() || (n.attr.Mode & os.ModeSymlink) != 0 {
// 		return nil, fmt.Errorf("Node is not a file.")
// 	}
//
// 	file, ok := n.Node.(*File)
// 	if !ok {
// 		return nil, fmt.Errorf("Node was recorded to be a file, but is not.")
// 	}
//
// 	return file, nil
// }

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

func (n *Node) UpdateTimes(updates UpdateTime) {
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

// func (n *Node) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
// 	oldParent, err := n.fs.GetNode(n.parent)
// 	if err != nil {
// 		return err
// 	}
//
// 	oldDir, err := oldParent.Dir()
// 	if err != nil {
// 		return err
// 	}
//
// 	newParentInum := Inum(req.NewDir)
//
// 	if newParentInum == n.parent {
// 		oldDir.children[req.NewName] = n
// 		delete(oldDir.children, req.OldName)
// 	} else {
// 		parent, err := n.fs.GetNode(newParentInum)
// 		if err != nil {
// 			return err
// 		}
//
// 		dir, err := parent.Dir()
// 		if err != nil {
// 			return err
// 		}
//
// 		dir.children[req.NewName] = n
// 		delete(oldDir.children, req.OldName)
// 	}
//
// 	return nil
// }
