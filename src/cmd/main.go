package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"sync"
	"time"
	"top"
	otelSetup "top/otel"

	"github.com/prometheus/procfs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdown, err := otelSetup.SetupOTelSDK(ctx)
	if err != nil {
		panic(err)
	}
	defer shutdown(ctx)

	var meter = otel.Meter("top")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	fs, err := procfs.NewFS("/proc")
	if err != nil {
		logger.Error("failed to open proc", "err", err)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	t := top.NewTop(&fs)
	processMap := make(map[int]*top.ProcessTop)

	userCPU, _ := meter.Float64Gauge("node.cpu.user")
	idleCPU, _ := meter.Float64Gauge("node.cpu.idle")
	systemCPU, _ := meter.Float64Gauge("node.cpu.system")

	processCPU, _ := meter.Float64Gauge("process.cpu.usage")
	processMemory, _ := meter.Float64Gauge("process.memory.usage")

	fmt.Println("start mytop")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer func() {
			for _, proc := range processMap {
				proc.Cpu = 0
				proc.Memory = 0

				attrs := []attribute.KeyValue{
					attribute.String("command", proc.Command)}

				processCPU.Record(ctx, proc.Cpu, metric.WithAttributes(attrs...))
				processMemory.Record(ctx, proc.Memory, metric.WithAttributes(attrs...))
			}
			wg.Done()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// リセットする
				for _, proc := range processMap {
					proc.Cpu = 0
					proc.Memory = 0
					processMap[proc.Pid] = proc
				}

				// topの値を更新します
				err := t.Update()
				if err != nil {
					logger.Error("failed to update top.proc", "err", err)
					continue
				}

				cpu, procs := t.Snapshot()

				userCPU.Record(ctx, cpu.User)
				idleCPU.Record(ctx, cpu.Idle)
				systemCPU.Record(ctx, cpu.System)

				for _, proc := range procs {
					processMap[proc.Pid] = proc
				}

				// 生きているプロセスのみ集計
				procGroup := make(map[string]*top.ProcessTop)
				for _, proc := range processMap {
					if proc.State != "R" {
						continue
					}
					// マルチプロセスの場合一つにまとめる
					if existing, ok := procGroup[proc.Command]; ok {
						existing.Cpu += proc.Cpu
						existing.Memory += proc.Memory
					} else {
						procGroup[proc.Command] = &top.ProcessTop{
							Command: proc.Command,
							Cpu:     proc.Cpu,
							Memory:  proc.Memory,
						}
					}
				}

				var processList []*top.ProcessTop
				for _, proc := range procGroup {
					processList = append(processList, proc)
				}

				sort.Slice(processList, func(i, j int) bool {
					return processList[i].Cpu > processList[j].Cpu
				})

				// デバッグ用
				//for _, proc := range processList {
				//	fmt.Printf("Command: %s, CPU: %.2f, Memory: %.2f, State: %s\n",
				//		proc.Command, proc.Cpu, proc.Memory, proc.State)
				//}

				// top10のみ送信
				for _, proc := range processList[:10] {
					attrs := []attribute.KeyValue{
						attribute.String("command", proc.Command)}

					processCPU.Record(ctx, proc.Cpu, metric.WithAttributes(attrs...))
					processMemory.Record(ctx, proc.Memory, metric.WithAttributes(attrs...))
				}

				// 死んだプロセスを削除
				for key, proc := range processMap {
					if proc.Cpu == 0 && proc.Memory == 0 {
						delete(processMap, key)
					}
				}
			}
		}
	}()
	wg.Wait()

}
