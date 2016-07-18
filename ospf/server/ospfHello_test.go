/*
OspfHello_test.go
hello packet testing routines.
*/

package server

import (
	"fmt"
	"l3/ospf/config"
	"testing"
)

func initHelloTestParams() {
	ospf = getServerObject()
	initAttr()
	go startDummyChannels(ospf)
}

func printHelloData(pkt OSPFHelloData) {
	fmt.Println("Netmask ", pkt.netmask)
	fmt.Println("HelloInterval ", pkt.helloInterval)
	fmt.Println("options ", pkt.options)
	fmt.Println("rtrPrio ", pkt.rtrPrio)
	fmt.Println("rtrDeadInt ", pkt.rtrDeadInterval)
	fmt.Println("designatedRtr ", pkt.designatedRtr)
	fmt.Println("backupDesignatedRtr ", pkt.backupDesignatedRtr)
}

func TestOspfHelloDecode(t *testing.T) {
	fmt.Println("\n\n**************** HELLO PROTOCOL ************\n")
	initHelloTestParams()
	for index := 1; index < 21; index++ {
		err := helloTestLogic(index)
		if err != SUCCESS {
			fmt.Println("Failed test case for Hello protocol ")
		}
	}

}

func helloTestLogic(tNum int) int {
	ospf.initDefaultIntfConf(key, ipIntfProp, ifType)
	data := []byte{0x01, 0x00, 0x5e, 0x00, 0x00, 0x05, 0xca, 0x11, 0x09, 0xb3, 0x00, 0x1c, 0x08, 0x00, 0x45, 0xc0,
		0x00, 0x50, 0x8d, 0xed, 0x00, 0x00, 0x01, 0x59, 0x3f, 0x5a, 0x0a, 0x4b, 0x00, 0xfe, 0xe0, 0x00, 0x00, 0x05, 0x02, 0x01, 0x00, 0x30, 0x4b, 0x01, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00, 0x3e, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x00, 0x0a, 0x12, 0x01, 0x00, 0x00, 0x00, 0x28, 0x0a, 0x4b, 0x00, 0xfe, 0x0a, 0x4b, 0x00, 0x01, 0x4b, 0x01, 0x00, 0x01, 0xff, 0xf6, 0x00, 0x03, 0x00, 0x01, 0x00, 0x04, 0x00, 0x00, 0x00, 0x01}

	ospf.IntfConfMap[key] = intf
	// start the go routine to receive messages from hello protocol
	go startDummyIntfChannels(key)
	//start tests

	switch tNum {

	case 1:
		decodeData := NewOSPFHelloData()
		decodeOspfHelloData(data, decodeData)

	case 2:
		fmt.Println(tNum, ": Running decodeOspfHelloData with corrupt data ")
		data = []byte{0x01, 0x00, 0x5e, 0x00, 0x00, 0x05, 0xca, 0x11, 0x09, 0xb3, 0x00, 0x1c, 0x08, 0x00, 0x45, 0xc0,
			0x00, 0x50, 0x8d, 0xed, 0x00, 0x00, 0x01, 0x59, 0x3f, 0x5a, 0x0a, 0x4b, 0x00, 0xfe, 0xe0, 0x00, 0x00, 0x05, 0x02, 0x01, 0x00, 0x30, 0x4b, 0x01, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00, 0x3e, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x00, 0x0a, 0x12, 0x01, 0x00, 0x00, 0x00, 0x28, 0x0a, 0x4b, 0x00, 0xfe, 0x0a, 0x4b, 0x00, 0x01, 0x4b, 0x02, 0x00, 0x01, 0xff, 0xf6, 0x00, 0x03, 0x00, 0x01, 0x00, 0x04, 0x00, 0x00, 0x00, 0x01, 0x02, 0x02, 0x02}
		decodeData := NewOSPFHelloData()
		decodeOspfHelloData(data, decodeData)
		printHelloData(*decodeData)

	case 3:
		fmt.Println(tNum, ": Running encodeOspfHelloData")
		encode := encodeOspfHelloData(hellodata)
		fmt.Println("Encoded data : ", encode)

	case 4:
		fmt.Println(tNum, ": Running BuildHelloPkt")
		pkt := ospf.BuildHelloPkt(intf)
		fmt.Println(" Encoded packet : ", pkt)

	case 5:
		fmt.Println(tNum, ": Running processRxHelloPkt ")
		err := ospf.processRxHelloPkt(hello, &ospfHdrMd, &ipHdrMd, &ethHdrMd, key)
		if err != nil {
			fmt.Println("Failed to process received hello data ", err)
		}

	case 6:
		fmt.Println(tNum, ": Running processOspfHelloNeighbor")
		//ospf.processOspfHelloNeighbor(false, &hellodata, &ipHdrMd, &ospfHdrMd, key)

	case 7:
		fmt.Println(tNum, ": Running CreateAndSendHelloRecvdMsg ")
		ospf.CreateAndSendHelloRecvdMsg(12, &ipHdrMd, &ospfHdrMd, 40, config.Broadcast, true, 1, key)

	case 8:
		fmt.Println(tNum, ": Running CreateAndSendHelloRecvdMsg ")
		ospf.CreateAndSendHelloRecvdMsg(12, &ipHdrMd, &ospfHdrMd, 40, config.Broadcast, false, 1, key)
	case 9:
		fmt.Println(tNum, ": Running header APIs.")
		checkHeaderAPIs()
	}
	return SUCCESS
}

func checkHeaderAPIs() {

	ospf.IntfConfMap[key] = intf
	intf.SendMutex = sync.Mutex{}

	ospfHdr := &OSPFHeader{}
	decodeOspfHdr(header, ospfHdr)

	pkt := encodeOspfHdr(*ospfHdr)
	fmt.Println("Encoded header pkt : ", pkt)

	ospf.processOspfHeader(hello, key, &ospfHdrMd)
	ospf.processOspfData(hello, &ethHdrMd, &ipHdrMd, &ospfHdrMd, key)
	ospf.processOspfData(lsaupd, &ethHdrMd, &ipHdrMd, &ospfHdrMd, key)
	ospf.processOspfData(lsaack, &ethHdrMd, &ipHdrMd, &ospfHdrMd, key)
	ospf.processOspfData(lsareq, &ethHdrMd, &ipHdrMd, &ospfHdrMd, key)

	ospf.processIPv4Layer(ip_layer, dstIP, &ipHdrMd)

	ospf.StopSendHelloPkt(key)
	ospf.StartSendHelloPkt(key)

	ospf.SendOspfPkt(key, hello)
}
