package top

import (
	"github.com/prometheus/procfs"
)

type ProcessManager struct {
	currentProcess  map[int]*Process
	previousProcess map[int]*Process
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		currentProcess:  make(map[int]*Process),
		previousProcess: make(map[int]*Process),
	}
}

func (p *ProcessManager) Update(fs *procfs.FS) error {
	p.previousProcess = make(map[int]*Process)
	for _, proc := range p.currentProcess {
		p.previousProcess[proc.pid] = proc
	}
	p.currentProcess = make(map[int]*Process)

	procs, err := fs.AllProcs()
	if err != nil {
		return err
	}

	for _, proc := range procs {
		stat, err := proc.Stat()
		if err != nil {
			return err
		}
		procInfo := NewProcess(stat.PID, stat.Comm, stat.State, stat.UTime, stat.STime, stat.VSize, stat.ResidentMemory()/1024)
		p.currentProcess[procInfo.pid] = procInfo
	}

	return nil
}

type ProcessTop struct {
	Pid     int
	Command string
	State   string
	uTime   uint
	sTime   uint
	Vs      uint // virtual memory size
	Rss     int  // resident set size
	Cpu     float64
	Memory  float64
}

type Process struct {
	pid     int
	command string
	state   string
	uTime   uint
	sTime   uint
	vs      uint // virtual memory size
	rss     int  // resident set size
}

func NewProcess(pid int, command string, state string, uTime uint, sTime uint, vs uint, rss int) *Process {
	return &Process{
		pid:     pid,
		command: command,
		state:   state,
		uTime:   uTime,
		sTime:   sTime,
		rss:     rss,
		vs:      vs,
	}
}

func (p *Process) Sub(previousProcess *Process) *Process {
	if previousProcess == nil {
		return p
	}
	return &Process{
		pid:     p.pid,
		command: p.command,
		state:   p.state,
		uTime:   p.uTime - previousProcess.uTime, // processは累計で取得しているのでdeltaをとる
		sTime:   p.sTime - previousProcess.sTime,
		rss:     p.rss,
		vs:      p.vs,
	}
}
