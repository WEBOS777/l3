Copyright [2016] [SnapRoute Inc]

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

	 Unless required by applicable law or agreed to in writing, software
	 distributed under the License is distributed on an "AS IS" BASIS,
	 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	 See the License for the specific language governing permissions and
	 limitations under the License.
// ribNotify.go
package server

import (
	"fmt"
	"github.com/op/go-nanomsg"
	"time"
)

type NotificationMsg struct {
	pub_socket *nanomsg.PubSocket
	msg        []byte
	eventInfo  string
}

func (ribdServiceHandler *RIBDServer) NotificationServer() {
	logger.Info(fmt.Sprintln("Starting notification server loop"))
	for {
		notificationMsg := <-ribdServiceHandler.NotificationChannel
		logger.Info(fmt.Sprintln("Event received with eventInfo: ", notificationMsg.eventInfo))
		eventInfo := RouteEventInfo{timeStamp: time.Now().String(), eventInfo: notificationMsg.eventInfo}
		localRouteEventsDB = append(localRouteEventsDB, eventInfo)
		notificationMsg.pub_socket.Send(notificationMsg.msg, nanomsg.DontWait)
	}
}
