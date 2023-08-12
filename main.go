package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	_ "net/http/pprof"

	"github.com/chenliu1993/simple-csi-driver/internal/nfs"
	"github.com/chenliu1993/simple-csi-driver/pkg/utils"
	"k8s.io/klog/v2"
)

const (
	// General Plugin Name
	TypePluginName = "plugin.csi.cliufreever.com"
	// NFS CSI NAME
	TypePluginNFS = "nfsplugin.csi.cliufreever.com"
)

var (
	endpoint = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	driver   = flag.String("driver", TypePluginNFS, "name of the driver")
	nodeName = flag.String("node", "", "node name")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	if *nodeName == "" {
		klog.Warning("Node name is required.")
	}

	// For debugging
	pprofPort := os.Getenv("PPROF_PORT")
	if pprofPort != "" {
		if _, err := strconv.Atoi(pprofPort); err == nil {
			klog.V(2).InfoS("Enable pprof at: %s", pprofPort)
			go func() {
				err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", pprofPort), nil)
				klog.V(2).ErrorS(err, "Start pprof error")
			}()
		}
	}

	klog.V(2).Infof("Driver %s is running at %s on node %s", *driver, *endpoint, *nodeName)

	driverList := *driver
	drivers := strings.Split(driverList, ",")
	// Start the relevant drivers
	var wg sync.WaitGroup

	for idx, d := range drivers {
		if !strings.Contains(d, TypePluginName) {
			drivers[idx] = d + "." + TypePluginName
		}
	}

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGTERM)
	stopChs := utils.GenMultiChs(drivers, stopCh)
	go utils.SendToMultiChs(stopChs, stopCh)

	for _, driver := range drivers {
		wg.Add(1)
		klog.V(2).Infof("CSI endpoint for driver %s: %s", driver, *endpoint)

		switch driver {
		case TypePluginNFS:
			go func(endpoint string) {
				defer wg.Done()
				nfsDriver := nfs.NewNFSDriver(TypePluginNFS, endpoint, *nodeName, stopChs[TypePluginNFS])
				nfsDriver.Run()
			}(*endpoint)
		}
	}

	wg.Wait()
	klog.V(2).Infof("Driver %s stopped", *driver)
	os.Exit(0)
}
