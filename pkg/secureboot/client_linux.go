//go:build linux

package secureboot

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	teeIOCVersion      = 0x800ca400
	teeIOCOpenSession  = 0x8010a402
	teeIOCInvoke       = 0x8010a403
	teeIOCCloseSession = 0x8004a405
	teeIOCShmAlloc     = 0xc010a401

	teeImplIDOPTEE = 1

	teeLoginPublic = 0

	teeAttrTypeValueInput   = 1
	teeAttrTypeMemrefInput  = 5
	teeAttrTypeMemrefOutput = 6

	ptaRkSecureBootGetInfo        = 0x0
	ptaRkSecureBootBurnHash       = 0x1
	ptaRkSecureBootLockdownDevice = 0x2
)

var ptaRkSecureBootUUID = [16]byte{
	0x5c, 0xfa, 0x57, 0xf6, 0x1a, 0x4c, 0x40, 0x7f,
	0x94, 0xa7, 0xa5, 0x6c, 0x8c, 0x47, 0x01, 0x9d,
}

type teeVersionData struct {
	ImplID   uint32
	ImplCaps uint32
	GenCaps  uint32
}

type teeShmAllocData struct {
	Size  uint64
	Flags uint32
	ID    int32
}

type teeBufData struct {
	BufPtr uint64
	BufLen uint64
}

type teeOpenSessionArg struct {
	UUID      [16]byte
	ClntUUID  [16]byte
	ClntLogin uint32
	CancelID  uint32
	Session   uint32
	Ret       uint32
	RetOrigin uint32
	NumParams uint32
}

type teeInvokeArg struct {
	Func      uint32
	Session   uint32
	CancelID  uint32
	Ret       uint32
	RetOrigin uint32
	NumParams uint32
}

type teeCloseSessionArg struct {
	Session uint32
}

type teeParam struct {
	Attr uint64
	A    uint64
	B    uint64
	C    uint64
}

type Info struct {
	Enabled    bool
	Simulation bool
	Hash       [32]byte
}

type Client struct {
	fd      int
	session uint32
}

func NewClient() (*Client, error) {
	fd, err := openTEEDevice()
	if err != nil {
		return nil, err
	}

	client := &Client{fd: fd}
	if err := client.openSession(); err != nil {
		_ = unix.Close(fd)
		return nil, err
	}

	return client, nil
}

func (c *Client) Close() {
	if c.fd <= 0 {
		return
	}

	closeArg := teeCloseSessionArg{Session: c.session}
	_ = ioctlPtr(c.fd, teeIOCCloseSession, uintptr(unsafe.Pointer(&closeArg)))
	_ = unix.Close(c.fd)
	c.fd = -1
}

func (c *Client) openSession() error {
	n := int(unsafe.Sizeof(teeOpenSessionArg{}))
	buf := make([]byte, n)
	arg := (*teeOpenSessionArg)(unsafe.Pointer(&buf[0]))
	arg.UUID = ptaRkSecureBootUUID
	arg.ClntLogin = teeLoginPublic
	arg.NumParams = 0

	bufData := teeBufData{
		BufPtr: uint64(uintptr(unsafe.Pointer(&buf[0]))),
		BufLen: uint64(len(buf)),
	}

	if err := ioctlPtr(c.fd, teeIOCOpenSession, uintptr(unsafe.Pointer(&bufData))); err != nil {
		return err
	}
	if arg.Ret != 0 {
		return fmt.Errorf("open session failed: ret=0x%x origin=0x%x", arg.Ret, arg.RetOrigin)
	}

	c.session = arg.Session
	return nil
}

func (c *Client) GetInfo() (Info, error) {
	var out Info

	shm, err := allocSHM(c.fd, 34)
	if err != nil {
		return out, err
	}
	defer shm.Close()

	params := []teeParam{
		{
			Attr: teeAttrTypeMemrefOutput,
			A:    0,
			B:    uint64(len(shm.Data)),
			C:    uint64(uint32(shm.ID)),
		},
	}

	ret, origin, err := c.invoke(ptaRkSecureBootGetInfo, params)
	if err != nil {
		return out, err
	}
	if ret != 0 {
		return out, fmt.Errorf("get_info failed: ret=0x%x origin=0x%x", ret, origin)
	}

	out.Enabled = shm.Data[0] != 0
	out.Simulation = shm.Data[1] != 0
	copy(out.Hash[:], shm.Data[2:34])
	return out, nil
}

func (c *Client) BurnHash(hash [32]byte, keySize uint32) error {
	shm, err := allocSHM(c.fd, uint64(len(hash)))
	if err != nil {
		return err
	}
	defer shm.Close()

	copy(shm.Data, hash[:])

	params := []teeParam{
		{
			Attr: teeAttrTypeMemrefInput,
			A:    0,
			B:    uint64(len(hash)),
			C:    uint64(uint32(shm.ID)),
		},
		{
			Attr: teeAttrTypeValueInput,
			A:    uint64(keySize),
			B:    0,
			C:    0,
		},
	}

	ret, origin, err := c.invoke(ptaRkSecureBootBurnHash, params)
	if err != nil {
		return err
	}
	if ret != 0 {
		return fmt.Errorf("burn_hash failed: ret=0x%x origin=0x%x", ret, origin)
	}
	return nil
}

func (c *Client) LockdownDevice() error {
	ret, origin, err := c.invoke(ptaRkSecureBootLockdownDevice, nil)
	if err != nil {
		return err
	}
	if ret != 0 {
		return fmt.Errorf("lockdown_device failed: ret=0x%x origin=0x%x", ret, origin)
	}
	return nil
}

func (c *Client) invoke(function uint32, params []teeParam) (uint32, uint32, error) {
	headerSize := int(unsafe.Sizeof(teeInvokeArg{}))
	paramSize := int(unsafe.Sizeof(teeParam{}))
	totalSize := headerSize + len(params)*paramSize

	buf := make([]byte, totalSize)
	arg := (*teeInvokeArg)(unsafe.Pointer(&buf[0]))
	arg.Func = function
	arg.Session = c.session
	arg.NumParams = uint32(len(params))

	if len(params) > 0 {
		paramPtr := unsafe.Pointer(uintptr(unsafe.Pointer(&buf[0])) + uintptr(headerSize))
		copy(unsafe.Slice((*teeParam)(paramPtr), len(params)), params)
	}

	bufData := teeBufData{
		BufPtr: uint64(uintptr(unsafe.Pointer(&buf[0]))),
		BufLen: uint64(len(buf)),
	}

	if err := ioctlPtr(c.fd, teeIOCInvoke, uintptr(unsafe.Pointer(&bufData))); err != nil {
		return 0, 0, err
	}
	return arg.Ret, arg.RetOrigin, nil
}

type teeSharedMemory struct {
	FD   int
	ID   int32
	Data []byte
}

func (m *teeSharedMemory) Close() {
	if m == nil {
		return
	}
	if len(m.Data) > 0 {
		_ = unix.Munmap(m.Data)
	}
	if m.FD > 0 {
		_ = unix.Close(m.FD)
	}
}

func allocSHM(teeFD int, size uint64) (*teeSharedMemory, error) {
	alloc := teeShmAllocData{Size: size}
	shmFD, err := ioctlPtrRetFD(teeFD, teeIOCShmAlloc, uintptr(unsafe.Pointer(&alloc)))
	if err != nil {
		return nil, err
	}

	data, err := unix.Mmap(shmFD, 0, int(size), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		_ = unix.Close(shmFD)
		return nil, err
	}

	return &teeSharedMemory{
		FD:   shmFD,
		ID:   alloc.ID,
		Data: data,
	}, nil
}

func openTEEDevice() (int, error) {
	paths := []string{"/dev/teepriv0", "/dev/tee0"}
	var lastErr error

	for _, p := range paths {
		fd, err := unix.Open(p, os.O_RDWR, 0)
		if err != nil {
			lastErr = err
			continue
		}

		v := teeVersionData{}
		if err := ioctlPtr(fd, teeIOCVersion, uintptr(unsafe.Pointer(&v))); err != nil {
			_ = unix.Close(fd)
			lastErr = err
			continue
		}

		if v.ImplID != teeImplIDOPTEE {
			_ = unix.Close(fd)
			lastErr = fmt.Errorf("%s is not an OP-TEE device (impl_id=%d)", p, v.ImplID)
			continue
		}

		return fd, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no TEE device found")
	}
	return -1, lastErr
}

func ioctlPtr(fd int, req uintptr, ptr uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), req, ptr)
	if errno != 0 {
		return errno
	}
	return nil
}

func ioctlPtrRetFD(fd int, req uintptr, ptr uintptr) (int, error) {
	ret, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), req, ptr)
	if errno != 0 {
		return 0, errno
	}
	return int(ret), nil
}
