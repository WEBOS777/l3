//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//       Unless required by applicable law or agreed to in writing, software
//       distributed under the License is distributed on an "AS IS" BASIS,
//       WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//       See the License for the specific language governing permissions and
//       limitations under the License.
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
	"infra/sysd/sysdCommonDefs"
	"log/syslog"
	"utils/dbutils"
	"utils/logging"
)

var routeServer *RIBDServer

func RIBdNewLogger(name string, tag string) (*logging.Writer, error) {
	var err error
	srLogger := new(logging.Writer)
	srLogger.MyComponentName = name

	srLogger.SysLogger, err = syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, tag)
	if err != nil {
		fmt.Println("Failed to initialize syslog - ", err)
		return srLogger, err
	}

	srLogger.GlobalLogging = true
	srLogger.MyLogLevel = sysdCommonDefs.DEBUG
	return srLogger, err
}

func getServerObject() *RIBDServer {
	logger, err := RIBdNewLogger("ribd", "RIBDTEST")
	if err != nil {
		fmt.Println("ribdtest: creating logger failed")
	}
	dbHdl := dbutils.NewDBUtil(logger)
	err = dbHdl.Connect()
	if err != nil {
		logger.Err("Failed to dial out to Redis server")
		return nil
	}
	server := NewRIBDServicesHandler(dbHdl, logger)
	if server == nil {
		fmt.Sprintln("ribd server object is null ")
		return nil
	}
	routeServer = server
	return server
}
