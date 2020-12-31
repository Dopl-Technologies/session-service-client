package client

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/grpc"

	dtprotos "github.com/dopl-technologies/api-protos-go"
)

// Client device service client
type Client struct {
	client dtprotos.SessionServiceClient
	conn   *grpc.ClientConn
}

// New creates a new client that connects to the
// given address
func New(address string) (Interface, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &Client{
		client: dtprotos.NewSessionServiceClient(conn),
		conn:   conn,
	}, nil
}

// Create creates a new session
func (c *Client) Create(name string, deviceIDs []uint64) (*dtprotos.Session, error) {
	req := &dtprotos.CreateSessionRequest{
		Name:      name,
		DeviceIDs: deviceIDs,
	}
	res, err := c.client.Create(context.Background(), req)
	if err != nil {
		return nil, err
	}

	session := res.GetSession()
	if session == nil {
		return nil, fmt.Errorf("unexpected session in create response. Status is ok but session is nil")
	}

	return session, nil
}

// Get gets a session by its id
func (c *Client) Get(id uint64) (*dtprotos.Session, error) {
	req := &dtprotos.GetSessionRequest{
		SessionID: id,
	}
	res, err := c.client.Get(context.Background(), req)
	if err != nil {
		return nil, err
	}

	session := res.GetSession()
	if session == nil {
		return nil, fmt.Errorf("unexpected get session response. Status is ok but session is nil")
	}

	return session, nil
}

// List lists sessions
func (c *Client) List() ([]*dtprotos.Session, error) {
	req := &dtprotos.ListSessionsRequest{}
	stream, err := c.client.List(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// Build the list
	var sessions []*dtprotos.Session
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, res.GetSession())
	}

	return sessions, nil
}

// Delete deletes a session by its id
func (c *Client) Delete(id uint64) error {
	req := &dtprotos.DeleteSessionRequest{
		SessionID: id,
	}
	_, err := c.client.Delete(context.Background(), req)
	return err
}

// WaitFor maintains a connection to the server and
// updates the returned channel with a session that is ready to join
func (c *Client) WaitFor(deviceID uint64) (<-chan *dtprotos.WaitForSessionResponse, context.CancelFunc, error) {
	req := &dtprotos.WaitForSessionRequest{
		DeviceID: deviceID,
	}
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := c.client.WaitFor(ctx, req)
	if err != nil {
		return nil, cancel, err
	}

	respCh := make(chan *dtprotos.WaitForSessionResponse)
	go func() {
		for {
			res, err := stream.Recv()
			if err != nil {
				close(respCh)
				break
			}

			respCh <- res
		}
	}()

	return respCh, cancel, nil
}

// ListWaiting lists devices that are waiting for a session
func (c *Client) ListWaiting() ([]*dtprotos.Device, error) {
	req := &dtprotos.ListWaitingSessionRequest{}
	stream, err := c.client.ListWaiting(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// Build the list
	var devices []*dtprotos.Device
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		devices = append(devices, res.GetDevice())
	}

	return devices, nil
}

// Join joins the session and returns a channel that's updated with session info
func (c *Client) Join(deviceID uint64, sessionID uint64) (<-chan *dtprotos.SessionDevice, context.CancelFunc, error) {
	req := &dtprotos.JoinSessionRequest{
		DeviceID:  deviceID,
		SessionID: sessionID,
	}
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := c.client.Join(ctx, req)
	if err != nil {
		return nil, cancel, err
	}

	// Create a channel that handles responses
	ch := make(chan *dtprotos.SessionDevice)
	go func() {
		for {
			res, err := stream.Recv()
			if err != nil {
				close(ch)
				break
			}

			ch <- res.GetSessionDevice()
		}
	}()

	return ch, cancel, nil
}

// Close closes the connection
func (c *Client) Close() {
	c.conn.Close()
}
