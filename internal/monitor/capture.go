package monitor

import (
    "fmt"
    "time"

    "github.com/google/gopacket"
    "github.com/google/gopacket/pcap"
    log "github.com/sirupsen/logrus"
)

type Capture struct {
    handles map[string]*pcap.Handle
}

func NewCapture() *Capture {
    return &Capture{
        handles: make(map[string]*pcap.Handle),
    }
}

func (c *Capture) Start(containerName string) error {
    iface := fmt.Sprintf("veth%s", containerName)
    handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)
    if err != nil {
        return fmt.Errorf("failed to open capture: %w", err)
    }

    if err := handle.SetBPFFilter("tcp"); err != nil {
        handle.Close()
        return fmt.Errorf("failed to set filter: %w", err)
    }

    c.handles[containerName] = handle

    go c.capture(containerName, handle)
    return nil
}

func (c *Capture) Stop(containerName string) {
    if handle, exists := c.handles[containerName]; exists {
        handle.Close()
        delete(c.handles, containerName)
    }
}

func (c *Capture) capture(containerName string, handle *pcap.Handle) {
    source := gopacket.NewPacketSource(handle, handle.LinkType())
    for packet := range source.Packets() {
        log.WithFields(log.Fields{
            "container": containerName,
            "timestamp": time.Now(),
            "length":    len(packet.Data()),
            "protocol": packet.NetworkLayer().NetworkFlow().String(),
        }).Info("Captured packet")
    }
}
