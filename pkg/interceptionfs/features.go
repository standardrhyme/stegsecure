/*
 * This file describes the fuse features to include, with static checks to
 * ensure they were implemeneted correctly.
 */

package interceptionfs

import (
	"bazil.org/fuse/fs"
)

// FS

var _ = fs.FS(&FS{})

// Dir

var _ = Node(&Dir{})
var _ = fs.Node(&Dir{})
var _ = fs.NodeRequestLookuper(&Dir{})
var _ = fs.HandleReadDirAller(&Dir{})
var _ = fs.NodeMkdirer(&Dir{})
var _ = fs.NodeCreater(&Dir{})

// var _ = fs.NodeSymlinker(&Dir{})
// var _ = fs.NodeReadlinker(&Dir{})
// var _ = fs.NodeLinker(&Dir{})
var _ = fs.NodeRenamer(&Dir{})
var _ = fs.NodeRemover(&Dir{})

// File

var _ = Node(&File{})
var _ = fs.Node(&File{})

// var _ = fs.NodeReadlinker(&File{})
var _ = fs.NodeOpener(&File{})

// FileHandle

var _ = fs.Handle(&FileHandle{})
var _ = fs.HandleReader(&FileHandle{})
var _ = fs.HandleWriter(&FileHandle{})
