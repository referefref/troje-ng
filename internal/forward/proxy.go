package forward

import (
    "context"
    "fmt"
    "io"
    "net"
    "time"

    "github.com/username/troje/internal/container"
    "github.com/username/troje/internal/monitor"
    log "github.com/sirupsen/logrus"
)

type Proxy struct {
    manager *container.Manager
    capture *monitor.Capture
}

func NewProxy(mgr *container.Manager, cap *monitor.Capture) *Proxy {
    return &Proxy{
        manager: mgr,
        capture: cap,
    }
}

func (p *Proxy) Start(ctx context.Context, addr string) error {
    l, err := net.Listen("tcp", addr)
    if err != nil {
        return fmt.Errorf("failed to listen: %w", err)
    }
    defer l.Close()

    go func() {
        <-ctx.Done()
        l.Close()
    }()

    for {
        conn, err := l.Accept()
        if err != nil {
            if ctx.Err() != nil {
                return nil
            }
            log.Errorf("Accept failed: %v", err)
            continue
        }

        go p.handleConnection(ctx, conn)
    }
}

func (p *Proxy) handleConnection(ctx context.Context, clientConn net.Conn) {
    defer clientConn.Close()

    remoteIP := clientConn.RemoteAddr().(*net.TCPAddr).IP.String()
    log.Infof("New connection from %s", remoteIP)

    container, err := p.manager.GetContainer(remoteIP)
    if err != nil {
        log.Errorf("Failed to get container: %v", err)
        return
    }

    containerIP, err := container.IPAddress("eth0")
    if err != nil {
        log.Errorf("Failed to get container IP: %v", err)
        return
    }

    serverConn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:22", containerIP[0]), 10*time.Second)
    if err != nil {
        log.Errorf("Failed to connect to container: %v", err)
        return
    }
    defer serverConn.Close()

    // Start packet capture
    if err := p.capture.Start(container.Name()); err != nil {
        log.Errorf("Failed to start packet capture: %v", err)
    }
    defer p.capture.Stop(container.Name())

    // Bidirectional copy
    errCh := make(chan error, 2)
    go func() {
        _, err := io.Copy(serverConn, clientConn)
        errCh <- err
    }()
    go func() {
        _, err := io.Copy(clientConn, serverConn)
        errCh <- err
    }()

    // Wait for either connection to close
    err = <-errCh
    if err != nil && err != io.EOF {
        log.Errorf("Connection error: %v", err)
    }
}
