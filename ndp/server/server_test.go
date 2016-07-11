//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//	 Unless required by applicable law or agreed to in writing, software
//	 distributed under the License is distributed on an "AS IS" BASIS,
//	 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	 See the License for the specific language governing permissions and
//	 limitations under the License.
//
// _______  __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  |
// |  |__   |  |     |  |__   \  V  /     |   (----` \   \/    \/   /  |  | `---|  |----`|  ,----'|  |__|  |
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |     |  |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |     |  `----.|  |  |  |
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//
package server

import (
	"fmt"
	"github.com/google/gopacket/pcap"
	"infra/sysd/sysdCommonDefs"
	"l3/ndp/debug"
	"log/syslog"
	"testing"
	"utils/logging"
)

func NDPTestNewLogger(name string, tag string, listenToConfig bool) (*logging.Writer, error) {
	var err error
	srLogger := new(logging.Writer)
	srLogger.MyComponentName = name

	srLogger.SysLogger, err = syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, tag)
	if err != nil {
		fmt.Println("Failed to initialize syslog - ", err)
		return srLogger, err
	}

	srLogger.GlobalLogging = true
	srLogger.MyLogLevel = sysdCommonDefs.INFO
	return srLogger, err
}

// Test ND Solicitation message Decoder
func TestInvalidInitPortInfo(t *testing.T) {
	svr := &NDPServer{}
	svr.InitGlobalDS()

	if len(svr.PhyPort) > 0 {
		t.Error("There should not be any elements in the system port map", len(svr.PhyPort))
	}
	svr.DeInitGlobalDS()

	if svr.PhyPort != nil {
		t.Error("De-Init for ndp port info didn't happen")
	}
}

// Test ND Solicitation message Decoder
func TestInvalidInitL3Info(t *testing.T) {
	svr := &NDPServer{}
	svr.InitGlobalDS()
	svr.InitSystemIPIntf(nil, nil)

	if len(svr.L3Port) > 0 {
		t.Error("There should not be any elements in the system ip map", len(svr.L3Port))
	}
	svr.DeInitGlobalDS()

	if svr.L3Port != nil {
		t.Error("De-Init for ndp l3 info didn't happen")
	}
}

// Test Pcap Create
func TestPcapCreate(t *testing.T) {
	var err error
	var pcapHdl *pcap.Handle
	logger, err := NDPTestNewLogger("ndpd", "NDPTEST", true)
	if err != nil {
		t.Error("creating logger failed")
	}
	debug.NDPSetLogger(logger)
	svr := NDPNewServer(nil)
	svr.InitGlobalDS()
	pcapHdl, err = svr.CreatePcapHandler("lo")
	if err != nil {
		t.Error("Pcap Create Failed", err)
	}
	svr.DeletePcapHandler(pcapHdl)
	pcapHdl = nil
	svr.DeletePcapHandler(pcapHdl)
}
