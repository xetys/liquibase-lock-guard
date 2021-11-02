package main

import (
	"github.com/xetys/liquibase-lock-guard/pkg"
	"k8s.io/klog/v2"
	"time"
)

var DEBUG_STR = pkg.Getenv("DEBUG", "false")
var DEBUG = false

func checkCycle() {
	if DEBUG {
		klog.Infoln("starting check cycle")
	}
	postgrespods, err := pkg.GetPostgresPods()
	if err != nil {
		klog.Errorf("failed retrieving pods with error %s\n", err)
	}

	for _, postgrespod := range postgrespods {
		if DEBUG {
			klog.Infof("found postgres pod with name %s\n", postgrespod.Name)
		}
		isPodLocked, err := pkg.CheckPodForLock(&postgrespod)
		if err != nil {
			klog.Errorln(err)
		}

		if isPodLocked {
			// repair
			klog.Warningf("expired liquibase lock detected in %s...executing update\n", postgrespod.Name)
			err := pkg.ResetLiquibaseLock(&postgrespod)
			if err != nil {
				klog.Errorln(err)
			}
		}
	}
}

func main() {
	if DEBUG_STR == "true" {
		DEBUG = true
	}
	for {
		checkCycle()
		time.Sleep(30 * time.Second)
	}
}
