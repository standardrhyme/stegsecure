# stegSecure

stegSecure is a software that runs alongside existing antivirus solutions to intercept new user files upon download, scans the images for 
LSB-steganography, and santizes them of hidden information.

_**This version of stegSecure must be run on the Linux OS **_

## How to Run 
#### Step 1: Clone Git Repository
Clone the following git repository with `git clone https://github.com/standardrhyme/stegsecure`.

#### Step 2: Begin stegSecure 
Change the current directory into the recently cloned `stegsecure` folder. Start stegSecure with `sudo go run main.go`.

#### Step 3: Download an image 
Download an image from an Internet browser.

#### Step 4: To shut stegSecure down
Run `sudo umount`. 

## Options

**[-datapath]**
Specifies the directory to mount over.

## Screenshots

#### Running stegSecure without arguments:
![image](https://user-images.githubusercontent.com/15258611/146490512-059e2f48-a331-49b0-9e8b-5bcb0b29b063.png)

#### Running stegSecure with specified directory: 
![image](https://user-images.githubusercontent.com/15258611/146490285-fa9c339a-05b1-45e5-8569-bfd2281752a2.png)


## Diagrams

#### General Workflow
![image](https://user-images.githubusercontent.com/15258611/146490879-f082af56-f9eb-4796-a78e-4132164469ba.png)

#### File Interception Layers
![image](https://user-images.githubusercontent.com/15258611/146491332-b8787b4d-27c8-4314-a511-b8d13567b77e.png)

#### File Sanitation via Bit Masking
![image](https://user-images.githubusercontent.com/15258611/146491529-186164d2-1f7c-4061-b92d-b556101fcb94.png)


## Features
![image](https://user-images.githubusercontent.com/15258611/146491151-80a38b3f-a729-4902-837d-90a2defa54bb.png)


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

