// policyUtils.go
package server

import (
	"errors"
	"fmt"
	"l3/bgp/utils"
	"net"
	"strconv"
	"strings"
	"utils/patriciaDB"
	utilspolicy "utils/policy"
)

const (
	Add = iota
	Del
	DelAll
	Invalidate
)

const (
	Invalid = -1
	Valid   = 0
)

type RouteParams struct {
	DestNetIp  string
	PrefixLen  uint16
	NextHopIp  string
	CreateType int
	DeleteType int
}

type PolicyRouteIndex struct {
	DestNetIP string //CIDR format
	Policy    string
}

type localDB struct {
	prefix     patriciaDB.Prefix
	isValid    bool
	precedence int
	nextHopIp  string
}

type ConditionsAndActionsList struct {
	conditionList []string
	actionList    []string
}
type PolicyStmtMap struct {
	policyStmtMap map[string]ConditionsAndActionsList
}

var PolicyRouteMap map[PolicyRouteIndex]PolicyStmtMap

func getPolicyEnityKey(entity utilspolicy.PolicyEngineFilterEntityParams, policy string) (
	policyEntityKey utilspolicy.PolicyEntityMapIndex) {
	utils.Logger.Info(fmt.Sprintln("getPolicyEnityKey entity =", entity, "policy =", policy))
	policyEntityKey = PolicyRouteIndex{DestNetIP: entity.DestNetIp, Policy: policy}
	return policyEntityKey
}

func getIPInt(ip net.IP) (ipInt int, err error) {
	if ip == nil {
		fmt.Printf("ip address %v invalid\n", ip)
		return ipInt, errors.New("Invalid destination network IP Address")
	}
	ip = ip.To4()
	parsedPrefixIP := int(ip[3]) | int(ip[2])<<8 | int(ip[1])<<16 | int(ip[0])<<24
	ipInt = parsedPrefixIP
	return ipInt, nil
}

func getIP(ipAddr string) (ip net.IP, err error) {
	ip = net.ParseIP(ipAddr)
	if ip == nil {
		return ip, errors.New("Invalid destination network IP Address")
	}
	ip = ip.To4()
	return ip, nil
}

func getPrefixLen(networkMask net.IP) (prefixLen int, err error) {
	ipInt, err := getIPInt(networkMask)
	if err != nil {
		return -1, err
	}
	for prefixLen = 0; ipInt != 0; ipInt >>= 1 {
		prefixLen += ipInt & 1
	}
	return prefixLen, nil
}
func getNetworkPrefix(destNetIp net.IP, networkMask net.IP) (destNet patriciaDB.Prefix, err error) {
	prefixLen, err := getPrefixLen(networkMask)
	if err != nil {
		utils.Logger.Info(fmt.Sprintln("err when getting prefixLen, err= ", err))
		return destNet, err
	}
	vdestMask := net.IPv4Mask(networkMask[0], networkMask[1], networkMask[2], networkMask[3])
	netIp := destNetIp.Mask(vdestMask)
	numbytes := prefixLen / 8
	if (prefixLen % 8) != 0 {
		numbytes++
	}
	destNet = make([]byte, numbytes)
	for i := 0; i < numbytes; i++ {
		destNet[i] = netIp[i]
	}
	return destNet, err
}
func getNetowrkPrefixFromStrings(ipAddr string, mask string) (prefix patriciaDB.Prefix, err error) {
	destNetIpAddr, err := getIP(ipAddr)
	if err != nil {
		utils.Logger.Info(fmt.Sprintln("destNetIpAddr invalid"))
		return prefix, err
	}
	networkMaskAddr, err := getIP(mask)
	if err != nil {
		utils.Logger.Info(fmt.Sprintln("networkMaskAddr invalid"))
		return prefix, err
	}
	prefix, err = getNetworkPrefix(destNetIpAddr, networkMaskAddr)
	if err != nil {
		utils.Logger.Info(fmt.Sprintln("err=", err))
		return prefix, err
	}
	return prefix, err
}

func GetNetworkPrefixFromCIDR(ipAddr string) (ipPrefix patriciaDB.Prefix, err error) {
	//var ipMask net.IP
	_, ipNet, err := net.ParseCIDR(ipAddr)
	if err != nil {
		return ipPrefix, err
	}
	/*
		ipMask = make(net.IP, 4)
		copy(ipMask, ipNet.Mask)
		ipAddrStr := ip.String()
		ipMaskStr := net.IP(ipMask).String()
		ipPrefix, err = getNetowrkPrefixFromStrings(ipAddrStr, ipMaskStr)
	*/
	i := strings.IndexByte(ipAddr, '/')
	prefixLen, _ := strconv.Atoi(ipAddr[i+1:])
	numbytes := (prefixLen + 7) / 8
	destNet := make([]byte, numbytes)
	for i := 0; i < numbytes; i++ {
		destNet[i] = ipNet.IP[i]
	}

	return patriciaDB.Prefix(destNet), err
}

func (eng *BGPPolicyEngine) DeleteRoutePolicyState(route *Route, policyName string) {
	utils.Logger.Info(fmt.Sprintln("deleteRoutePolicyState"))
	found := false
	idx := 0
	/*    if routeInfoRecordList.policyList[policyName] != nil {
		delete(routeInfoRecordList.policyList, policyName)
	}*/
	for idx = 0; idx < len(route.PolicyList); idx++ {
		if route.PolicyList[idx] == policyName {
			found = true
			break
		}
	}

	if !found {
		utils.Logger.Info(fmt.Sprintln("Policy ", policyName, "not found in policyList of route", route))
		return
	}

	route.PolicyList = append(route.PolicyList[:idx], route.PolicyList[idx+1:]...)
}

func deleteRoutePolicyStateAll(route *Route) {
	utils.Logger.Info(fmt.Sprintln("deleteRoutePolicyStateAll"))
	route.PolicyList = nil
	return
}

func deletePolicyRouteMapEntry(route *Route, policy string) {
	utils.Logger.Info(fmt.Sprintln("deletePolicyRouteMapEntry for policy ", policy, "route ", route.bgpRoute.Network, "/",
		route.bgpRoute.CIDRLen))
	if PolicyRouteMap == nil {
		utils.Logger.Info(fmt.Sprintln("PolicyRouteMap empty"))
		return
	}
	destNetIP := route.bgpRoute.Network + "/" + strconv.Itoa(int(route.bgpRoute.CIDRLen))
	policyRouteIndex := PolicyRouteIndex{DestNetIP: destNetIP, Policy: policy}
	//PolicyRouteMap[policyRouteIndex].policyStmtMap=nil
	delete(PolicyRouteMap, policyRouteIndex)
}

func addRoutePolicyState(route *Route, policy string, policyStmt string) {
	utils.Logger.Info(fmt.Sprintln("addRoutePolicyState"))
	route.PolicyList = append(route.PolicyList, policy)
	return
}

func UpdateRoutePolicyState(route *Route, op int, policy string, policyStmt string) {
	utils.Logger.Info(fmt.Sprintln("updateRoutePolicyState"))
	if op == DelAll {
		deleteRoutePolicyStateAll(route)
		//deletePolicyRouteMapEntry(route, policy)
	} else if op == Add {
		addRoutePolicyState(route, policy, policyStmt)
	}
}

func (eng *BGPPolicyEngine) addPolicyRouteMap(route *Route, policy string) {
	utils.Logger.Info(fmt.Sprintln("addPolicyRouteMap"))
	//policy.hitCounter++
	//ipPrefix, err := getNetowrkPrefixFromStrings(route.Network, route.Mask)
	var newRoute string
	newRoute = route.bgpRoute.Network + "/" + strconv.Itoa(int(route.bgpRoute.CIDRLen))
	ipPrefix, err := GetNetworkPrefixFromCIDR(newRoute)
	if err != nil {
		utils.Logger.Info(fmt.Sprintln("Invalid ip prefix"))
		return
	}
	//  newRoute := string(ipPrefix[:])
	utils.Logger.Info(fmt.Sprintln("Adding ip prefix %s %v ", newRoute, ipPrefix))
	policyInfo := eng.PolicyEngine.PolicyDB.Get(patriciaDB.Prefix(policy))
	if policyInfo == nil {
		utils.Logger.Info(fmt.Sprintln("Unexpected:policyInfo nil for policy ", policy))
		return
	}
	tempPolicy := policyInfo.(utilspolicy.Policy)
	policyExtensions := tempPolicy.Extensions.(PolicyExtensions)
	policyExtensions.HitCounter++

	utils.Logger.Info(fmt.Sprintln("routelist len= ", len(policyExtensions.RouteList), " prefix list so far"))
	found := false
	for i := 0; i < len(policyExtensions.RouteList); i++ {
		utils.Logger.Info(fmt.Sprintln(policyExtensions.RouteList[i]))
		if policyExtensions.RouteList[i] == newRoute {
			utils.Logger.Info(fmt.Sprintln(newRoute, " already is a part of ", policy, "'s routelist"))
			found = true
		}
	}
	if !found {
		policyExtensions.RouteList = append(policyExtensions.RouteList, newRoute)
	}

	found = false
	utils.Logger.Info(fmt.Sprintln("routeInfoList details"))
	for i := 0; i < len(policyExtensions.RouteInfoList); i++ {
		utils.Logger.Info(fmt.Sprintln("IP: ", policyExtensions.RouteInfoList[i].bgpRoute.Network, "/",
			policyExtensions.RouteInfoList[i].bgpRoute.CIDRLen, " nextHop: ",
			policyExtensions.RouteInfoList[i].bgpRoute.NextHop))
		if policyExtensions.RouteInfoList[i].bgpRoute.Network == route.bgpRoute.Network &&
			policyExtensions.RouteInfoList[i].bgpRoute.CIDRLen == route.bgpRoute.CIDRLen &&
			policyExtensions.RouteInfoList[i].bgpRoute.NextHop == route.bgpRoute.NextHop {
			utils.Logger.Info(fmt.Sprintln("route already is a part of ", policy, "'s routeInfolist"))
			found = true
		}
	}
	if found == false {
		policyExtensions.RouteInfoList = append(policyExtensions.RouteInfoList, route)
	}
	eng.PolicyEngine.PolicyDB.Set(patriciaDB.Prefix(policy), tempPolicy)
}

func deletePolicyRouteMap(route *Route, policy string) {
	fmt.Println("deletePolicyRouteMap")
}

func (eng *BGPPolicyEngine) UpdatePolicyRouteMap(route *Route, policy string, op int) {
	utils.Logger.Info(fmt.Sprintln("updatePolicyRouteMap"))
	if op == Add {
		eng.addPolicyRouteMap(route, policy)
	} else if op == Del {
		deletePolicyRouteMap(route, policy)
	}

}