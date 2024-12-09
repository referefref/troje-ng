package main

import (
    "context"
    "flag"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"

    "github.com/username/troje/internal/container"
    "github.com/username/troje/internal/forward"
    "github.com/username/troje/internal/monitor"
    log "github.com/sirupsen/logrus"
)

var (
    baseContainer = flag.String("b", "", "Base container name")
    listenAddr   = flag.String("l", ":8022", "Listen address")
    maxIdle      = flag.Duration("idle", 1*time.Hour, "Maximum container idle time")
    emailTo      = flag.String("email", "", "Email address for notifications")
)

func main() {
    flag.Parse()

    if *baseContainer == "" {
        log.Fatal("No base container defined")
    }

    // Setup logging
    log.SetFormatter(&log.JSONFormatter{})
    log.SetOutput(os.Stdout)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Initialize container manager
    mgr, err := container.NewManager(*baseContainer)
    if err != nil {
        log.Fatalf("Failed to initialize container manager: %v", err)
    }

    // Initialize proxy server
    proxy := forward.NewProxy(mgr, monitor.NewCapture())

    // Start housekeeping
    go mgr.Housekeeping(ctx, *maxIdle)

    // Start proxy server
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := proxy.Start(ctx, *listenAddr); err != nil {
            log.Fatalf("Proxy server failed: %v", err)
        }
    }()

    // Handle shutdown
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    <-sigCh
    log.Info("Shutting down...")
    cancel()
    wg.Wait()
    mgr.Cleanup()
}
