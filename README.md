# stegsecure

stegSecure is a software that runs alongside existing antivirus solutions to intercept new user files upon download, scans the images for 
LSB-steganography, and santizes them of hidden information.

*** This version of stegSecure must be run on the Linux OS *** 

## How to Run 
#### Step 1: Clone Git Repository
Clone the following git repository with `git clone https://github.com/standardrhyme/stegsecure`.

#### Step 2: Begin stegSecure 
Change the current directory into the recently cloned `stegsecure` folder. Start stegSecure with `sudo go run main.go`.

#### Step 3: Download an image 
Download an image from an Internet browser.

## Screenshots

## Workflow

## Custom Data Structures
```go

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
```
## Exit Codes 
- `0`: Successful
- `1`: Incorrect command line input format
- `2`: External package function error

