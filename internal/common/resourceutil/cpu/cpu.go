package cpu

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lf-edge/edge-home-orchestration-go/internal/common/logmgr"
)

const (
	// USER is the key for user. Value set to 0
	USER = iota
	// NICE is the key for Nice, Value set to 1
	NICE
	// SYSTEM  is the key for system, Value set to 2
	SYSTEM
	// IDLE is the key for idle, Value set to 3
	IDLE
	// IOWAIT is the key for IO wait, Value set to 4
	IOWAIT
	// IRQ is the key for IRQ, Value set to 5
	IRQ
	// SOFTIRQ is the key for SOFTIRQ, Value set to 6
	SOFTIRQ
	// STEAL is the key for steal, Value set to 7
	STEAL
	// GEUST is the key for guest, Value set to 8
	GEUST
	// GESTNICE is the key for guest nice, Value set to 9
	GESTNICE
)

const (
	// CPUIdle is the key idle cpu. value set to 0
	CPUIdle = iota
	// CPUNonidle is key for non idle cpu
	CPUNonidle
	// CPUTotal is the key for total cpu value
	CPUTotal
)

//InfoStat struct
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
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer f.Close()

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
			cpuUsage[CPUNonidle] += item
		case IDLE, IOWAIT:
			cpuUsage[CPUIdle] += item
		}

	}

	cpuUsage[CPUTotal] = cpuUsage[CPUIdle] + cpuUsage[CPUNonidle]

	return cpuUsage, err
}

// Percent is used to return the cpu usage in percentage
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

	totald := curUsage[CPUTotal] - preUsage[CPUTotal]
	idled := curUsage[CPUIdle] - preUsage[CPUIdle]

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

// Info is used to return the frequency and core info in InfoStat structure
func Info() ([]InfoStat, error) {
	ret, err := getCPUs()
	if err != nil {
		return nil, err
	}

	feq, err := getCPUMaxFreq()
	if err != nil {
		feq, err = getCPUFreqCPUInfo()
		if err != nil {
			return nil, err
		}
	}

	ret[0].Mhz = feq

	return ret, nil
}

func getCPUMaxFreq() (float64, error) {
	f, err := fileOpen("/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq")
	if err != nil {
		return 0.0, err
	}
	defer f.Close()

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

func getCPUFreqCPUInfo() (float64, error) {
	f, err := fileOpen("/proc/cpuinfo")
	if err != nil {
		return 0.0, err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	var line string
	for {
		line, err = r.ReadString('\n')
		if err != nil {
			log.Error(err.Error())
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
	if err != nil {
		return nil, err
	}
	defer f.Close()

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
