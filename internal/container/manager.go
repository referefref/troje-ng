package container

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/lxc/go-lxc"
    log "github.com/sirupsen/logrus"
)

type Manager struct {
    baseContainer string
    containers    map[string]*Container
    mu           sync.RWMutex
}

type Container struct {
    *lxc.Container
    LastAccess time.Time
}

func NewManager(base string) (*Manager, error) {
    return &Manager{
        baseContainer: base,
        containers:    make(map[string]*Container),
    }, nil
}

func (m *Manager) GetContainer(ip string) (*Container, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    if c, exists := m.containers[ip]; exists {
        c.LastAccess = time.Now()
        return c, nil
    }

    name := fmt.Sprintf("troje_%s_%d", randomString(8), time.Now().Unix())
    base, err := lxc.NewContainer(m.baseContainer, lxc.DefaultConfigPath())
    if err != nil {
        return nil, fmt.Errorf("failed to open base container: %w", err)
    }
    defer lxc.PutContainer(base)

    if err := base.CloneUsing(name, lxc.Aufs, lxc.CloneSnapshot); err != nil {
        return nil, fmt.Errorf("failed to clone container: %w", err)
    }

    c, err := lxc.NewContainer(name, lxc.DefaultConfigPath())
    if err != nil {
        return nil, fmt.Errorf("failed to create container: %w", err)
    }

    if err := c.Start(); err != nil {
        return nil, fmt.Errorf("failed to start container: %w", err)
    }

    if !c.Wait(lxc.RUNNING, 30) {
        return nil, fmt.Errorf("container failed to start")
    }

    container := &Container{
        Container:  c,
        LastAccess: time.Now(),
    }
    m.containers[ip] = container

    return container, nil
}

func (m *Manager) Housekeeping(ctx context.Context, maxIdle time.Duration) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.cleanupIdle(maxIdle)
        }
    }
}

func (m *Manager) cleanupIdle(maxIdle time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()

    now := time.Now()
    for ip, c := range m.containers {
        if now.Sub(c.LastAccess) > maxIdle {
            log.Infof("Cleaning up idle container for IP %s", ip)
            if err := c.Stop(); err != nil {
                log.Errorf("Failed to stop container: %v", err)
            }
            if err := c.Destroy(); err != nil {
                log.Errorf("Failed to destroy container: %v", err)
            }
            delete(m.containers, ip)
        }
    }
}

func (m *Manager) Cleanup() {
    m.mu.Lock()
    defer m.mu.Unlock()

    for ip, c := range m.containers {
        log.Infof("Cleaning up container for IP %s", ip)
        if err := c.Stop(); err != nil {
            log.Errorf("Failed to stop container: %v", err)
        }
        if err := c.Destroy(); err != nil {
            log.Errorf("Failed to destroy container: %v", err)
        }
    }
    m.containers = make(map[string]*Container)
}
