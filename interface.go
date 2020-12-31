package client

import (
	"context"

	dtprotos "github.com/dopl-technologies/api-protos-go"
)

// Interface session service client interface
//go:generate moq -out client_mock.go . Interface
type Interface interface {
	Create(name string, deviceIDs []uint64) (*dtprotos.Session, error)

	Get(id uint64) (*dtprotos.Session, error)

	List() ([]*dtprotos.Session, error)

	Delete(id uint64) error

	WaitFor(deviceID uint64) (<-chan *dtprotos.WaitForSessionResponse, context.CancelFunc, error)

	ListWaiting() ([]*dtprotos.Device, error)

	Join(deviceID uint64, sessionID uint64) (<-chan *dtprotos.SessionDevice, context.CancelFunc, error)

	Close()
}
