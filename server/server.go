package main

import (
	"flag"
	"sync"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/ailidani/paxi"
	"github.com/ailidani/paxi/abd"
	"github.com/ailidani/paxi/blockchain"
	"github.com/ailidani/paxi/chain"
	"github.com/ailidani/paxi/dynamo"
	"github.com/ailidani/paxi/epaxos"
	"github.com/ailidani/paxi/hpaxos"
	"github.com/ailidani/paxi/kpaxos"
	"github.com/ailidani/paxi/log"
	"github.com/ailidani/paxi/m2paxos"
	"github.com/ailidani/paxi/paxos"
	"github.com/ailidani/paxi/paxos_group"
	"github.com/ailidani/paxi/sdpaxos"
	"github.com/ailidani/paxi/vpaxos"
	"github.com/ailidani/paxi/wankeeper"
	"github.com/ailidani/paxi/wpaxos"
)

var algorithm = flag.String("algorithm", "paxos", "Distributed algorithm")
var id = flag.String("id", "", "ID in format of Zone.Node.")
var simulation = flag.Bool("sim", false, "simulation mode")

var master = flag.String("master", "", "Master address.")

func replica(id paxi.ID) {
	if *master != "" {
		paxi.ConnectToMaster(*master, false, id)
	}

	log.Infof("node %v starting...", id)

	switch *algorithm {

	case "paxos":
		paxos.NewReplica(id).Run()

	case "epaxos":
		epaxos.NewReplica(id).Run()

	case "sdpaxos":
		sdpaxos.NewReplica(id).Run()

	case "wpaxos":
		wpaxos.NewReplica(id).Run()

	case "abd":
		abd.NewReplica(id).Run()

	case "chain":
		chain.NewReplica(id).Run()

	case "vpaxos":
		vpaxos.NewReplica(id).Run()

	case "wankeeper":
		wankeeper.NewReplica(id).Run()

	case "kpaxos":
		kpaxos.NewReplica(id).Run()

	case "paxos_groups":
		paxos_group.NewReplica(id).Run()

	case "dynamo":
		dynamo.NewReplica(id).Run()

	case "blockchain":
		blockchain.NewMiner(id).Run()

	case "m2paxos":
		m2paxos.NewReplica(id).Run()

	case "hpaxos":
		hpaxos.NewReplica(id).Run()

	default:
		panic("Unknown algorithm")
	}
}

func getCpuSample() (idle uint64, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		log.Errorf("node %v proc stat read %v ", id, err)
		return
	}
	lines := strings.Split(string(contents), "\n")
	//cpu user nice system idle io irq softirq
	for _, line := range(lines) {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			for i := 1; i < len(fields); i++ {
				time, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					log.Errorf("node %v Atoi error %s ", id, line)
				}
				total += time
				if i == 4 {
					idle = time
				}
			}
			return
		}
	}
	return
}

func cpuUtilization(id paxi.ID) {
	idlePrev, totalPrev := getCpuSample()
	for{
		time.Sleep(1 * time.Second)
		idleCurr, totalCurr := getCpuSample()
		idleTime  := float64(idleCurr - idlePrev)
		totalTime := float64(totalCurr - totalPrev)
		cpuUtilization := 100 * (totalTime - idleTime)/totalTime
		idlePrev, totalPrev = idleCurr, totalCurr
		log.Infof("node %v CPU utilization: %f ", id, cpuUtilization)
	}
}

func main() {
	paxi.Init()

	if *simulation {
		var wg sync.WaitGroup
		wg.Add(1)
		paxi.Simulation()
		for id := range paxi.GetConfig().Addrs {
			n := id
			go replica(n)
		}
		wg.Wait()
	} else {
		go cpuUtilization(paxi.ID(*id))
		replica(paxi.ID(*id))	
	}
}
