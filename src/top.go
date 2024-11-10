package top

import (
	"github.com/prometheus/procfs"
)

type Top struct {
	fs      *procfs.FS
	node    *NodeManager
	process *ProcessManager
}

func NewTop(fs *procfs.FS) *Top {
	memory, err := fs.Meminfo()
	if err != nil {
		panic(err)
	}
	return &Top{
		fs:      fs,
		node:    NewNodeManager(*memory.MemTotal),
		process: NewProcessManager(),
	}
}

func (t *Top) Update() error {
	// nodeの更新
	err := t.node.Update(t.fs)
	if err != nil {
		return err
	}

	// プロセスの更新
	err = t.process.Update(t.fs)
	if err != nil {
		return err
	}

	return nil
}

func (t *Top) Snapshot() (*CPUStat, map[int]*ProcessTop) {
	snapshots := make(map[int]*ProcessTop)

	for _, proc := range t.process.currentProcess {
		diffProc := proc.Sub(t.process.previousProcess[proc.pid])
		snapshot := &ProcessTop{
			Pid:     proc.pid,
			Command: proc.command,
			State:   proc.state,
			uTime:   diffProc.uTime,
			sTime:   diffProc.sTime,
			Rss:     proc.rss,
			Vs:      proc.vs,
			Cpu:     float64(diffProc.uTime+diffProc.sTime) / 100 / t.node.TotalCPUUsage * 100,
			Memory:  float64(proc.rss) / float64(t.node.TotalMemory) * 100,
		}
		snapshots[proc.pid] = snapshot
	}

	return t.node.SnapshotCPU, snapshots
}
