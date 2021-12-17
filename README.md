# stegsecure

stegSecure is a software that runs alongside existing antivirus solutions to detect
LSB-steganography images among new user files.

## How to Run 
#### Step 1: Clone Git Repository
Clone the following git repository with `git clone https://github.com/standardrhyme/stegsecure`.

#### Step 2:

#### Step 3: 


## Screenshot

## Workflow

## Custom Data Structures
```Python

type FS struct {
	Debug      func(msg interface{})
	conn       *fuse.Conn
	mountpoint string
	binddir    string

	rootInum  Inum
	root      fs.Node
	nodes     map[Inum]*NodeAttr
	notifiers map[Inum]*func(Node)
	nextInum  Inum
	notifier  func(Node)

	nextPassInum Inum
	passNodes    map[Inum]*NodeAttr
}

type Node interface {
	fs.Node

	FS() *FS
	Inum() Inum
	Name() string
	SetName(newName string)
	Parent() *Dir
	SetParent(newDir *Dir)
	GetNode() (*NodeAttr, error)
	GetRelPath() string
	GetRealPath() string
	Passthrough() bool
}

type NodeAttr struct {
	fs   *FS
	attr fuse.Attr
}

type Dir struct {
	fs   *FS
	inum Inum

	name        string
	parent      *Dir
	children    map[string]Node
	passthrough bool
}

type File struct {
	fs   *FS
	inum Inum

	name   string
	parent *Dir
	data   []byte

	cleaned     bool
	passthrough bool
}

type FileHandle struct {
	*File
	passthroughHandle *os.File
}

## Exit Codes 
- `0`: Successful
- `1`: Incorrect command line input format
- `2`: External package function error

