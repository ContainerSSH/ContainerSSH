package agentforward

import (
	"errors"
	"fmt"
	"io"
	"sync"

	protocol "github.com/containerssh/libcontainerssh/agentprotocol"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/log"
)

type agentForward struct {
	lock            sync.Mutex
	reverseForwards map[string]*protocol.ForwardCtx
	nX11Channels    uint32
	x11Forward      *protocol.ForwardCtx
	directForward   *protocol.ForwardCtx
	logger          log.Logger
}

func NewAgentForward(
	logger log.Logger,
) AgentForward {
	return &agentForward{
		reverseForwards: make(map[string]*protocol.ForwardCtx),
		logger:          logger,
	}
}

func (f *agentForward) HasDirectAgent() bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.directForward != nil
}

func (f *agentForward) HasX11() bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.x11Forward != nil
}

func (f *agentForward) mangleForwardAddr(proto string, addr string, port uint32) string {
	switch proto {
	case "tcp":
		return fmt.Sprintf("tcp-%s:%d", addr, port)
	case "unix":
		return fmt.Sprintf("unix-%s", addr)
	default:
		panic(fmt.Errorf("Requested to mangle unknown protocol %s", proto))
	}
}

func serveConnection(log log.Logger, dst io.WriteCloser, src io.ReadCloser) {
	_, err := io.Copy(dst, src)
	if err != nil && errors.Is(err, io.EOF) {
		log.Warning("Connection error", err)
	}
	_ = dst.Close()
	_ = src.Close()
}

func (f *agentForward) serveX11(connChan chan *protocol.Connection, reverseHandler sshserver.ReverseForward) {
	for {
		agentConn, ok := <-connChan
		if !ok {
			return
		}

		details := agentConn.Details()

		f.lock.Lock()
		if f.nX11Channels == 0 {
			f.lock.Unlock()
			_ = agentConn.Reject()
			continue
		}
		f.lock.Unlock()

		forwardChannel, _, err := reverseHandler.NewChannelX11(details.OriginatorAddress, details.OriginatorPort)
		if err != nil {
			f.logger.Warning("Failed to open X11 forwarding channel")
			return
		}

		err = agentConn.Accept()
		if err != nil {
			return
		}

		go serveConnection(f.logger, forwardChannel, agentConn)
		go serveConnection(f.logger, agentConn, forwardChannel)
	}
}

func (f *agentForward) serveReverseForward(connChan chan *protocol.Connection, reverseHandler sshserver.ReverseForward) {
	for {
		agentConn, ok := <-connChan
		if !ok {
			f.logger.Info("Connection channel closed, ending forward")
			return
		}

		details := agentConn.Details()

		forwardChannel, _, err := reverseHandler.NewChannelTCP(details.ConnectedAddress, details.ConnectedPort, details.OriginatorAddress, details.OriginatorPort)
		if err != nil {
			f.logger.Warning("Failed to open X11 forwarding channel")
			return
		}

		err = agentConn.Accept()
		if err != nil {
			return
		}

		go serveConnection(f.logger, forwardChannel, agentConn)
		go serveConnection(f.logger, agentConn, forwardChannel)
	}
}

func (f *agentForward) serveReverseForwardUnix(connChan chan *protocol.Connection, reverseHandler sshserver.ReverseForward) {
	for {
		agentConn, ok := <-connChan
		if !ok {
			f.logger.Info("Connection channel closed, ending forward")
			return
		}

		details := agentConn.Details()

		forwardChannel, _, err := reverseHandler.NewChannelUnix(details.ConnectedAddress)
		if err != nil {
			f.logger.Warning("Failed to open X11 forwarding channel")
			return
		}

		err = agentConn.Accept()
		if err != nil {
			return
		}
		go serveConnection(f.logger, forwardChannel, agentConn)
		go serveConnection(f.logger, agentConn, forwardChannel)
	}
}

func (f *agentForward) setupX11(
	setupAgentCallback func() (io.Reader, io.Writer, error),
	logger log.Logger,
	singleConnection bool,
	proto string,
	cookie string,
	screen uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	if f.x11Forward != nil {
		return fmt.Errorf("X11 forwarding already setup")
	}
	fromAgent, toAgent, err := setupAgentCallback()
	if err != nil {
		return err
	}
	f.x11Forward = protocol.NewForwardCtx(fromAgent, toAgent, logger)

	screenstr := fmt.Sprintf("%d", screen)
	connChan, err := f.x11Forward.StartX11ForwardClient(singleConnection, screenstr, proto, cookie)
	if err != nil {
		return err
	}
	go f.serveX11(connChan, reverseHandler)
	return nil
}

func (f *agentForward) NewX11Forwarding(
	setupAgentCallback func() (io.Reader, io.Writer, error),
	logger log.Logger,
	singleConnection bool,
	proto string,
	cookie string,
	screen uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	if f.x11Forward == nil {
		err := f.setupX11(
			setupAgentCallback,
			logger,
			singleConnection,
			proto,
			cookie,
			screen,
			reverseHandler,
		)
		if err != nil {
			return err
		}
	}
	f.nX11Channels++
	return nil
}

func (f *agentForward) NewTCPReverseForwarding(
	setupAgentCallback func() (io.Reader, io.Writer, error),
	logger log.Logger,
	bindHost string,
	bindPort uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	key := f.mangleForwardAddr("tcp", bindHost, bindPort)
	if _, ok := f.reverseForwards[key]; ok {
		return fmt.Errorf("Forwarding already started for this host/port combo")
	}

	fromAgent, toAgent, err := setupAgentCallback()
	if err != nil {
		return err
	}

	f.reverseForwards[key] = protocol.NewForwardCtx(fromAgent, toAgent, logger)
	connChan, err := f.reverseForwards[key].StartReverseForwardClient(bindHost, bindPort, false)
	if err != nil {
		return err
	}

	go f.serveReverseForward(connChan, reverseHandler)

	return nil
}

func (f *agentForward) NewUnixReverseForwarding(
	setupAgentCallback func() (io.Reader, io.Writer, error),
	logger log.Logger,
	path string,
	reverseHandler sshserver.ReverseForward,
) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	key := f.mangleForwardAddr("unix", path, 0)
	if _, ok := f.reverseForwards[key]; ok {
		return fmt.Errorf("Forwarding already started for this socket")
	}

	fromAgent, toAgent, err := setupAgentCallback()
	if err != nil {
		return err
	}

	f.reverseForwards[key] = protocol.NewForwardCtx(fromAgent, toAgent, logger)
	connChan, err := f.reverseForwards[key].StartReverseForwardClientUnix(path, false)
	if err != nil {
		return err
	}

	go f.serveReverseForwardUnix(connChan, reverseHandler)

	return nil
}

func (f *agentForward) CancelTCPForwarding(
	bindHost string,
	bindPort uint32,
) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	key := f.mangleForwardAddr("tcp", bindHost, bindPort)
	ctx, ok := f.reverseForwards[key]
	if !ok {
		return fmt.Errorf("Forwarding not found")
	}
	err := ctx.NoMoreConnections()
	if err != nil {
		return err
	}
	delete(f.reverseForwards, key)

	return nil
}

func (f *agentForward) CancelStreamLocalForwarding(
	path string,
) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	key := f.mangleForwardAddr("unix", path, 0)
	ctx, ok := f.reverseForwards[key]
	if !ok {
		return fmt.Errorf("Forwarding not found")
	}
	err := ctx.NoMoreConnections()
	if err != nil {
		return err
	}
	delete(f.reverseForwards, key)

	return nil
}

func (f *agentForward) CloseX11Forwarding() error {
	f.lock.Lock()
	defer f.lock.Unlock()
	if f.x11Forward == nil {
		return fmt.Errorf("X11 forwarding not setup, please call NewX11Forwarding at least once")
	}
	if f.nX11Channels == 0 {
		return fmt.Errorf("Tried to close X11 session when there are already 0 X11 channels")
	}
	f.nX11Channels--

	return nil
}

func (f *agentForward) setupDirectForward(
	setupAgentCallback func() (io.Reader, io.Writer, error),
	logger log.Logger,
) error {
	if f.directForward != nil {
		return fmt.Errorf("Direct forwarding has already been setup for this connection")
	}
	fromAgent, toAgent, err := setupAgentCallback()
	if err != nil {
		return err
	}
	f.directForward = protocol.NewForwardCtx(fromAgent, toAgent, logger)
	connChan, err := f.directForward.StartServerForward()
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, ok := <-connChan
			if !ok {
				break
			}
			_ = conn.Reject()
		}
	}()
	return nil
}

func (f *agentForward) NewForwardTCP(
	setupAgentCallback func() (io.Reader, io.Writer, error),
	logger log.Logger,
	hostToConnect string,
	portToConnect uint32,
	originatorHost string,
	originatorPort uint32,
) (sshserver.ForwardChannel, error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	if f.directForward == nil {
		err := f.setupDirectForward(setupAgentCallback, logger)
		if err != nil {
			return nil, err
		}
	}
	conn, err := f.directForward.NewConnectionTCP(
		hostToConnect,
		portToConnect,
		originatorHost,
		originatorPort,
		nil,
	)
	return conn, err
}

func (f *agentForward) NewForwardUnix(
	setupAgentCallback func() (io.Reader, io.Writer, error),
	logger log.Logger,
	path string,
) (sshserver.ForwardChannel, error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	if f.directForward == nil {
		err := f.setupDirectForward(setupAgentCallback, logger)
		if err != nil {
			return nil, err
		}
	}
	conn, err := f.directForward.NewConnectionUnix(
		path,
		nil,
	)
	return conn, err
}

func (f *agentForward) OnShutdown() {
	f.lock.Lock()
	defer f.lock.Unlock()
	if f.directForward != nil {
		_ = f.directForward.NoMoreConnections()
		f.directForward.Kill()
	}
	if f.x11Forward != nil {
		_ = f.directForward.NoMoreConnections()
		f.x11Forward.Kill()
	}
	for _, forward := range f.reverseForwards {
		_ = forward.NoMoreConnections()
		forward.Kill()
	}
}
