//go:build !linux

package secureboot

import "fmt"

type Info struct {
	Enabled    bool
	Simulation bool
	Hash       [32]byte
}

type Client struct{}

func NewClient() (*Client, error) {
	return nil, fmt.Errorf("secure boot client is only supported on linux")
}

func (c *Client) Close() {}

func (c *Client) GetInfo() (Info, error) {
	return Info{}, fmt.Errorf("secure boot client is only supported on linux")
}

func (c *Client) BurnHash(hash [32]byte, keySize uint32) error {
	return fmt.Errorf("secure boot client is only supported on linux")
}

func (c *Client) LockdownDevice() error {
	return fmt.Errorf("secure boot client is only supported on linux")
}
