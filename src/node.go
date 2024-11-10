package top

import "github.com/prometheus/procfs"

type CPUStat struct {
	User      float64
	Nice      float64
	System    float64
	Idle      float64
	Iowait    float64
	IRQ       float64
	SoftIRQ   float64
	Steal     float64
	Guest     float64
	GuestNice float64
}

type NodeManager struct {
	previousCPU   procfs.CPUStat
	currentCPU    procfs.CPUStat
	SnapshotCPU   *CPUStat
	TotalMemory   uint64 // kilobyte
	TotalCPUUsage float64
}

func NewNodeManager(totalMemory uint64) *NodeManager {
	return &NodeManager{
		currentCPU:  procfs.CPUStat{},
		previousCPU: procfs.CPUStat{},
		TotalMemory: totalMemory,
	}
}

func (n *NodeManager) Update(fs *procfs.FS) error {
	stats, err := fs.Stat()
	if err != nil {
		return err
	}
	n.currentCPU = stats.CPUTotal

	userDelta := n.currentCPU.User - n.previousCPU.User
	niceDelta := n.currentCPU.Nice - n.previousCPU.Nice
	systemDelta := n.currentCPU.System - n.previousCPU.System
	idleDelta := n.currentCPU.Idle - n.previousCPU.Idle
	iowaitDelta := n.currentCPU.Iowait - n.previousCPU.Iowait
	irqDelta := n.currentCPU.IRQ - n.previousCPU.IRQ
	softIRQDelta := n.currentCPU.SoftIRQ - n.previousCPU.SoftIRQ
	stealDelta := n.currentCPU.Steal - n.previousCPU.Steal
	guestDelta := n.currentCPU.Guest - n.previousCPU.Guest
	guestNiceDelta := n.currentCPU.GuestNice - n.previousCPU.GuestNice

	total := userDelta + niceDelta + systemDelta + idleDelta + iowaitDelta + irqDelta + softIRQDelta + stealDelta + guestDelta + guestNiceDelta

	n.SnapshotCPU = &CPUStat{
		User:      (userDelta / total) * 100,
		Nice:      (niceDelta / total) * 100,
		System:    (systemDelta / total) * 100,
		Idle:      (idleDelta / total) * 100,
		Iowait:    (iowaitDelta / total) * 100,
		IRQ:       (irqDelta / total) * 100,
		SoftIRQ:   (softIRQDelta / total) * 100,
		Steal:     (stealDelta / total) * 100,
		Guest:     (guestDelta / total) * 100,
		GuestNice: (guestNiceDelta / total) * 100,
	}

	n.TotalCPUUsage = total
	n.previousCPU = n.currentCPU

	return nil
}
