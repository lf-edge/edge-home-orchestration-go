package cpu

import (
	"bufio"
	"github.com/lf-edge/edge-home-orchestration-go/src/common/logmgr"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	USER = iota
	NICE
	SYSTEM
	IDLE
	IOWAIT
	IRQ
	SOFTIRQ
	STEAL
	GEUST
	GESTNICE
)

const (
	CPU_IDLE = iota
	CPU_NONIDLE
	CPU_TOTAL
)

type InfoStat struct {
	Mhz  float64
	core int
}

var (
	fileOpen func(name string) (*os.File, error)
	log      = logmgr.GetInstance()
)

func init() {
	fileOpen = os.Open
}

func readCPUUsage() ([]float64, error) {
	f, err := fileOpen("/proc/stat")
	defer f.Close()
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	r := bufio.NewReader(f)
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	cpuUsage := make([]float64, 3)
	for index, field := range strings.Split(strings.Trim(line, "\n"), " ")[2:] {
		item, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return nil, err
		}

		switch index {
		case USER, NICE, SYSTEM, IRQ, SOFTIRQ, STEAL:
			cpuUsage[CPU_NONIDLE] += item
		case IDLE, IOWAIT:
			cpuUsage[CPU_IDLE] += item
		}

	}

	cpuUsage[CPU_TOTAL] = cpuUsage[CPU_IDLE] + cpuUsage[CPU_NONIDLE]

	return cpuUsage, err
}

func Percent(interval time.Duration, percpu bool) ([]float64, error) {
	preUsage, err := readCPUUsage()
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(1 * interval)
	<-ticker.C

	curUsage, err := readCPUUsage()
	if err != nil {
		return nil, err
	}

	totald := curUsage[CPU_TOTAL] - preUsage[CPU_TOTAL]
	idled := curUsage[CPU_IDLE] - preUsage[CPU_IDLE]

	cpuUsage := (totald - idled) / totald
	if cpuUsage < 0.0 {
		cpuUsage = 0.0
	}

	ticker.Stop()

	if percpu {
		cpuUsage *= 100
	}

	cpuUsages := make([]float64, 1)
	cpuUsages[0] = cpuUsage

	return cpuUsages, nil
}

func Info() ([]InfoStat, error) {
	ret, err := getCPUs()
	if err != nil {
		return nil, err
	}

	feq, err := getCPUMaxFreq()
	if err != nil {
		feq, err = getCPUFreqCpuInfo()
		if err != nil {
			return nil, err
		}
	}

	ret[0].Mhz = feq

	return ret, nil
}

func getCPUMaxFreq() (float64, error) {
	f, err := fileOpen("/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq")
	defer f.Close()
	if err != nil {
		return 0.0, err
	}

	strFeq, err := bufio.NewReader(f).ReadString('\n')
	if err != nil {
		return 0.0, err
	}

	feq, err := strconv.ParseFloat(strings.Trim(strFeq, "\n"), 64)
	if err != nil {
		return 0.0, err
	}

	for {
		if feq < 10000 {
			break
		}
		feq /= 1000.0
	}

	return feq, nil
}

func getCPUFreqCpuInfo() (float64, error) {
	f, err := fileOpen("/proc/cpuinfo")
	defer f.Close()
	if err != nil {
		return 0.0, err
	}

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}

		fields := strings.Split(strings.Trim(line, "\n"), ":")
		if len(fields) < 2 {
			continue
		}

		switch strings.TrimSpace(fields[0]) {
		//case "cpu MHz":
		//	feq, err := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
		//	if err != nil {
		//		return 0.0, err
		//	}
		//	return feq, nil
		case "model name":
			name := strings.Split(fields[1], " ")

			for _, val := range name {
				if strings.Contains(val, "GHz") {
					feq, err := strconv.ParseFloat(strings.Trim(val, "GHz"), 64)
					if err != nil {
							return 0.0, err
						}
					feq *= 1000.0
					return feq, nil
				}
			}
		}
	}

	return 0.0, err
}

func getCPUs() ([]InfoStat, error) {
	f, err := fileOpen("/proc/cpuinfo")
	defer f.Close()
	if err != nil {
		return nil, err
	}

	r := bufio.NewReader(f)
	ret := make([]InfoStat, 0)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}

		fields := strings.Split(strings.Trim(line, "\n"), ":")
		if len(fields) < 2 {
			continue
		}

		switch strings.TrimSpace(fields[0]) {
		case "processor":
			value, err := strconv.Atoi(strings.TrimSpace(fields[1]))
			if err != nil {
				continue
			}
			ret = append(ret, InfoStat{core: value})
		}
	}

	return ret, nil
}
