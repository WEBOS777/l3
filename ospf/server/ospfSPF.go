package server

import (
        "fmt"
        "errors"
        //"l3/ospf/config"
)

type VertexKey struct {
        Type      uint8
        ID        uint32
        AdvRtr          uint32
}

type Path []VertexKey

type TreeVertex struct {
        Paths           []Path
        Distance        uint16
        NumOfPaths      int
}

type StubVertex struct {
        NbrVertexKey    VertexKey
        NbrVertexCost   uint16
        LinkData        uint32
        LsaKey          LsaKey
        AreaId          uint32
        LinkStateId     uint32
}

type Vertex struct {
        NbrVertexKey    []VertexKey
        NbrVertexCost   []uint16
        LinkData        map[VertexKey]uint32
        LsaKey          LsaKey
        AreaId          uint32
        Visited         bool
        LinkStateId     uint32
        NetMask         uint32
}

const (
        RouterVertex uint8 = 0
        SNetworkVertex uint8 = 1 // Stub
        TNetworkVertex uint8 = 2 // Transit
)

func findSelfOrigRouterLsaKey(ent map[LsaKey]bool) (LsaKey, error) {
        var key LsaKey
        for key, _ := range ent {
                if key.LSType == RouterLSA {
                        return key, nil
                }
        }
        err := errors.New("No Self Orignated Router LSA found")
        return key, err
}

func (server *OSPFServer)UpdateAreaGraphNetworkLsa(lsaEnt NetworkLsa, lsaKey LsaKey, areaId uint32) error {
        server.logger.Info(fmt.Sprintln("2: Using Lsa with key as:", dumpLsaKey(lsaKey), "for SPF calc"))
        vertexKey := VertexKey {
                Type: TNetworkVertex,
                ID: lsaKey.LSId,
                AdvRtr: lsaKey.AdvRouter,
        }
        ent, exist := server.AreaGraph[vertexKey]
        if exist {
                server.logger.Info(fmt.Sprintln("Entry already exists in SPF Graph for vertexKey:", vertexKey))
                server.logger.Info(fmt.Sprintln("SPF Graph:", server.AreaGraph))
                return nil
        }
        ent.NbrVertexKey = make([]VertexKey, 0)
        ent.NbrVertexCost = make([]uint16, 0)
        ent.LinkData = make(map[VertexKey]uint32)
        for i := 0; i < len(lsaEnt.AttachedRtr); i++ {
                Rtr := lsaEnt.AttachedRtr[i]
                server.logger.Info(fmt.Sprintln("Attached Router at index:", i, "is:", Rtr))
                var vKey VertexKey
                var cost uint16
                vKey = VertexKey {
                        Type: RouterVertex,
                        ID: Rtr,
                        AdvRtr: Rtr,
                }
                cost = 0
                ent.NbrVertexKey = append(ent.NbrVertexKey, vKey)
                ent.NbrVertexCost = append(ent.NbrVertexCost, cost)
                ent.LinkData[vKey] = lsaEnt.Netmask
        }
        ent.AreaId = areaId
        ent.LsaKey = lsaKey
        ent.Visited = false
        //ent.NetMask = lsaEnt.NetMask
        ent.LinkStateId = lsaKey.LSId
        server.AreaGraph[vertexKey] = ent
        lsdbKey := LsdbKey{
                AreaId: areaId,
        }
        lsDbEnt, exist := server.AreaLsdb[lsdbKey]
        if !exist {
                server.logger.Err(fmt.Sprintln("No LS Database found for areaId:", areaId))
                err := errors.New(fmt.Sprintln("No LS Database found for areaId:", areaId))
                return err
        }
        for _, vKey := range ent.NbrVertexKey {
                _, exist := server.AreaGraph[vKey]
                if exist {
                        server.logger.Info(fmt.Sprintln("Entry already exists in SPF Graph for vertexKey:", vertexKey))
                        continue
                }
                lsaKey := LsaKey {
                        LSType: RouterLSA,
                        LSId: vKey.ID,
                        AdvRouter: vKey.AdvRtr,
                }
                lsaEnt, exist := lsDbEnt.RouterLsaMap[lsaKey]
                if !exist {
                        server.logger.Err(fmt.Sprintln("Router LSA with LsaKey:", lsaKey, "not found in areaId:", areaId))
                        server.logger.Err(fmt.Sprintln(lsDbEnt))
                        server.logger.Err(fmt.Sprintln("======Router LsaMap====", lsDbEnt.RouterLsaMap))
                        server.logger.Err(fmt.Sprintln("========Network LsaMap====", lsDbEnt.NetworkLsaMap))
                        err := errors.New(fmt.Sprintln("Router LSA with LsaKey:", lsaKey, "not found in areaId:", areaId))
                       // continue
                        return err
                }
                err := server.UpdateAreaGraphRouterLsa(lsaEnt, lsaKey, areaId)
                if err != nil {
                        return err
                }
        }
        return nil
}

func (server *OSPFServer)findNetworkLsa(areaId uint32, LSId uint32) (lsaKey LsaKey, err error) {
        lsdbKey := LsdbKey{
                AreaId: areaId,
        }
        lsDbEnt, exist := server.AreaLsdb[lsdbKey]
        if !exist {
                server.logger.Err(fmt.Sprintln("No LS Database found for areaId:", areaId))
                return
        }

        for key, _ := range lsDbEnt.NetworkLsaMap {
                if key.LSId == LSId &&
                key.LSType == NetworkLSA {
                        return key, nil
                }
        }

        err = errors.New("Network LSA not found")
        return lsaKey, err
}

func (server *OSPFServer)UpdateAreaGraphRouterLsa(lsaEnt RouterLsa, lsaKey LsaKey, areaId uint32) error {
        server.logger.Info(fmt.Sprintln("1: Using Lsa with key as:", dumpLsaKey(lsaKey), "for SPF calc"))
        vertexKey := VertexKey {
                Type: RouterVertex,
                ID: lsaKey.LSId,
                AdvRtr: lsaKey.AdvRouter,
        }
        ent, exist := server.AreaGraph[vertexKey]
        if exist {
                server.logger.Info(fmt.Sprintln("Entry already exists in SPF Graph for vertexKey:", vertexKey))
                server.logger.Info(fmt.Sprintln("SPF Graph:", server.AreaGraph))
                return nil
        }
        ent.NbrVertexKey = make([]VertexKey, 0)
        ent.NbrVertexCost = make([]uint16, 0)
        ent.LinkData = make(map[VertexKey]uint32)
        for i := 0; i < int(lsaEnt.NumOfLinks); i++ {
                server.logger.Info(fmt.Sprintln("Link Detail at index", i, "is:", lsaEnt.LinkDetails[i]))
                linkDetail := lsaEnt.LinkDetails[i]
                var vKey VertexKey
                var cost uint16
                var lData uint32
                if linkDetail.LinkType == TransitLink {
                        server.logger.Info("===It is TransitLink===")
                        vKey = VertexKey {
                                Type: TNetworkVertex,
                                ID: linkDetail.LinkId,
                                AdvRtr: 0,
                        }
                        nLsaKey, err := server.findNetworkLsa(areaId, vKey.ID)
                        if err != nil {
                                server.logger.Info(fmt.Sprintln("Err:",err))
                                return err
                                //continue
                        }
                        vKey.AdvRtr = nLsaKey.AdvRouter
                        cost = linkDetail.LinkMetric
                        lData = linkDetail.LinkData
                        ent.NbrVertexKey = append(ent.NbrVertexKey, vKey)
                        ent.NbrVertexCost = append(ent.NbrVertexCost, cost)
                        ent.LinkData[vKey] = lData
                } else if linkDetail.LinkType == StubLink {
                        server.logger.Info("===It is StubLink===")
                        vKey = VertexKey {
                                Type: SNetworkVertex,
                                ID: linkDetail.LinkId,
                                AdvRtr: lsaKey.AdvRouter,
                        }
                        cost = linkDetail.LinkMetric
                        lData = linkDetail.LinkData
                        sentry, _ := server.AreaStubs[vKey]
                        sentry.NbrVertexKey = vertexKey
                        sentry.NbrVertexCost = cost
                        sentry.LinkData = lData
                        sentry.AreaId = areaId
                        sentry.LsaKey = lsaKey
                        sentry.LinkStateId = lsaKey.LSId
                        server.AreaStubs[vKey] = sentry
                } else if linkDetail.LinkType == P2PLink {
                        // TODO
                }
        }
        ent.AreaId = areaId
        ent.LsaKey = lsaKey
        ent.Visited = false
        ent.LinkStateId = lsaKey.LSId
        server.AreaGraph[vertexKey] = ent
        lsdbKey := LsdbKey{
                AreaId: areaId,
        }
        lsDbEnt, exist := server.AreaLsdb[lsdbKey]
        if !exist {
                server.logger.Err(fmt.Sprintln("No LS Database found for areaId:", areaId))
                err := errors.New(fmt.Sprintln("No LS Database found for areaId:", areaId))
                return err
        }
        for _, vKey := range ent.NbrVertexKey {
                _, exist := server.AreaGraph[vKey]
                if exist {
                        server.logger.Info(fmt.Sprintln("Entry for Vertex:", vKey, "already exist in Area Graph"))
                        continue
                }
                lsaKey := LsaKey {
                        LSType: 0,
                        LSId: vKey.ID,
                        AdvRouter: vKey.AdvRtr,
                }
                if vKey.Type == TNetworkVertex {
                        lsaKey.LSType = NetworkLSA
                        lsaEnt, exist := lsDbEnt.NetworkLsaMap[lsaKey]
                        if !exist {
                                server.logger.Err(fmt.Sprintln("Network LSA with LsaKey:", lsaKey, "not found in LS Database of areaId:", areaId))
                                err := errors.New(fmt.Sprintln("Network LSA with LsaKey:", lsaKey, "not found in LS Database of areaId:", areaId))
                                //continue
                                return err
                        }
                        err := server.UpdateAreaGraphNetworkLsa(lsaEnt, lsaKey, areaId)
                        if err != nil {
                                return err
                        }
                } else if vKey.Type == RouterVertex {
                        lsaKey.LSType = RouterLSA
                        lsaEnt, exist := lsDbEnt.RouterLsaMap[lsaKey]
                        if !exist {
                                server.logger.Err(fmt.Sprintln("Router LSA with LsaKey:", lsaKey, "not found in LS Database of areaId:", areaId))
                                err := errors.New(fmt.Sprintln("Router LSA with LsaKey:", lsaKey, "not found in LS Database of areaId:", areaId))
                               //continue
                                return err
                        }
                        err := server.UpdateAreaGraphRouterLsa(lsaEnt, lsaKey, areaId)
                        if err != nil {
                                return err
                        }
                } else if vKey.Type == SNetworkVertex {
                        //TODO
                }
        }
        return nil
}

func (server *OSPFServer) CreateAreaGraph(areaId uint32) (VertexKey, error) {
        var vKey VertexKey
        server.logger.Info(fmt.Sprintln("Create SPF Graph for: areaId:", areaId))
        lsdbKey := LsdbKey{
                AreaId: areaId,
        }
        lsDbEnt, exist := server.AreaLsdb[lsdbKey]
        if !exist {
                server.logger.Err(fmt.Sprintln("No LS Database found for areaId:", areaId))
                err := errors.New(fmt.Sprintln("No LS Database found for areaId:", areaId))
                return vKey, err
        }

        selfOrigLsaEnt, exist := server.AreaSelfOrigLsa[lsdbKey]
        if !exist {
                server.logger.Err(fmt.Sprintln("No Self Originated LSAs found for areaId:", areaId))
                err := errors.New(fmt.Sprintln("No Self Originated LSAs found for areaId:", areaId))
                return vKey, err
        }
        selfRtrLsaKey, err := findSelfOrigRouterLsaKey(selfOrigLsaEnt)
        if err != nil {
                server.logger.Err(fmt.Sprintln("No Self Originated Router LSA Key found for areaId:", areaId))
                err := errors.New(fmt.Sprintln("No Self Originated Router LSA Key found for areaId:", areaId))
                return vKey, err
        }
        server.logger.Info(fmt.Sprintln("Self Orginated Router LSA Key:", selfRtrLsaKey))
        lsaEnt, exist := lsDbEnt.RouterLsaMap[selfRtrLsaKey]
        if !exist {
                server.logger.Err(fmt.Sprintln("No Self Originated Router LSA found for areaId:", areaId))
                err := errors.New(fmt.Sprintln("No Self Originated Router LSA found for areaId:", areaId))
                return vKey, err
        }

        err = server.UpdateAreaGraphRouterLsa(lsaEnt, selfRtrLsaKey, areaId)
        vKey = VertexKey {
                Type: RouterVertex,
                ID: selfRtrLsaKey.LSId,
                AdvRtr: selfRtrLsaKey.AdvRouter,
        }
        return vKey, err
}

func (server *OSPFServer)ExecuteDijkstra(vKey VertexKey, areaId uint32) error {
        var treeVSlice []VertexKey = make([]VertexKey, 0)

        treeVSlice = append(treeVSlice, vKey)
        ent, exist := server.SPFTree[vKey]
        if !exist {
                ent.Distance = 0
                ent.NumOfPaths = 1
                ent.Paths = make([]Path, 1)
                var path Path
                path = make(Path, 0)
                ent.Paths[0] = path
                server.SPFTree[vKey] = ent
        }
        for j := 0; j < len(treeVSlice); j++ {
                ent, exist := server.AreaGraph[treeVSlice[j]]
                if !exist {
                        server.logger.Info(fmt.Sprintln("No entry found for:", treeVSlice[j]))
                        err := errors.New(fmt.Sprintln("No entry found for:", treeVSlice[j]))
                        //continue
                        return err
                }
                for i := 0; i < len(ent.NbrVertexKey); i++ {
                        verKey := ent.NbrVertexKey[i]
                        cost := ent.NbrVertexCost[i]
                        entry, exist := server.AreaGraph[verKey]
                        if !exist {
                                server.logger.Info("Something is wrong in SPF Calculation: Entry should exist in Area Graph")
                                err := errors.New("Something is wrong in SPF Calculation: Entry should exist in Area Graph")
                                return err
                        } else {
                                if entry.Visited == true {
                                        continue
                                }
                        }
                        tEnt, exist := server.SPFTree[verKey]
                        if !exist {
                                tEnt.Paths = make([]Path, 1)
                                var path Path
                                path = make(Path, 0)
                                tEnt.Paths[0] = path
                                tEnt.Distance = 0xff00
                                tEnt.NumOfPaths = 1
                        }
                        tEntry, exist := server.SPFTree[treeVSlice[j]]
                        if !exist {
                                server.logger.Err("Something is wrong is SPF Calculation")
                                err := errors.New("Something is wrong is SPF Calculation")
                                return err
                        }
                        if tEnt.Distance > tEntry.Distance + cost {
                                tEnt.Distance = tEntry.Distance + cost
                                for l := 0; l < tEnt.NumOfPaths; l++ {
                                        tEnt.Paths[l] = nil
                                }
                                tEnt.Paths = tEnt.Paths[:0]
                                tEnt.NumOfPaths = 0
                                tEnt.Paths = nil
                                tEnt.Paths = make([]Path, tEntry.NumOfPaths)
                                for l := 0; l < tEntry.NumOfPaths; l++ {
                                        var path Path
                                        path = make(Path, len(tEntry.Paths[l]) + 1)
                                        copy(path, tEntry.Paths[l])
                                        path[len(tEntry.Paths[l])] = treeVSlice[j]
                                        tEnt.Paths[l] = path
                                }
                                tEnt.NumOfPaths = tEntry.NumOfPaths
                        } else if tEnt.Distance == tEntry.Distance + cost {
                                paths := make([]Path, (tEntry.NumOfPaths + tEnt.NumOfPaths))
                                for l := 0; l < tEnt.NumOfPaths; l++ {
                                        var path Path
                                        path = make(Path, len(tEnt.Paths[l]))
                                        copy(path, tEnt.Paths[l])
                                        paths[l] = path
                                        tEnt.Paths[l] = nil
                                }
                                tEnt.Paths = tEnt.Paths[:0]
                                tEnt.NumOfPaths = 0
                                tEnt.Paths = nil
                                for l := 0; l < tEntry.NumOfPaths; l++ {
                                        var path Path
                                        path = make(Path, len(tEntry.Paths[l]) + 1)
                                        copy(path, tEntry.Paths[l])
                                        path[len(tEntry.Paths[l])] = treeVSlice[j]
                                        paths[tEnt.NumOfPaths+l] =  path
                                }
                                tEnt.Paths = paths
                                tEnt.NumOfPaths = tEntry.NumOfPaths + tEnt.NumOfPaths
                        }
                        server.SPFTree[verKey] = tEnt
                        treeVSlice = append(treeVSlice, verKey)
                }
                ent.Visited = true
                server.AreaGraph[treeVSlice[j]] = ent
        }

        //Handling Stub Networks

        server.logger.Info("Handle Stub Networks")
        for key, entry := range server.AreaStubs {
                //Finding the Vertex(Router) to which this stub is connected to
                vertexKey := entry.NbrVertexKey
                parent, exist := server.SPFTree[vertexKey]
                if !exist {
                        continue
                }
                ent, _ := server.SPFTree[key]
                ent.Distance = parent.Distance + entry.NbrVertexCost
                ent.Paths = make([]Path, parent.NumOfPaths)
                for i := 0; i < parent.NumOfPaths; i++ {
                        var path Path
                        path = make(Path, len(parent.Paths[i]) + 1)
                        copy(path, parent.Paths[i])
                        path[len(parent.Paths[i])] = vertexKey
                        ent.Paths[i] = path
                }
                ent.NumOfPaths = parent.NumOfPaths
                server.SPFTree[key] = ent
        }
        return nil
}

func dumpVertexKey(key VertexKey) string {
        var Type string
        if key.Type == RouterVertex {
                Type = "Router"
        } else if key.Type == SNetworkVertex {
                Type = "Stub"
        } else if key.Type == TNetworkVertex {
                Type = "Transit"
        }
        ID := convertUint32ToIPv4(key.ID)
        AdvRtr := convertUint32ToIPv4(key.AdvRtr)
        return fmt.Sprintln("Vertex Key[Type:", Type, "ID:", ID, "AdvRtr:", AdvRtr)

}

func dumpLsaKey(key LsaKey) string {
        var Type string
        if key.LSType == RouterLSA {
                Type = "Router LSA"
        } else if key.LSType == NetworkLSA {
                Type = "Network LSA"
        }

        LSId := convertUint32ToIPv4(key.LSId)
        AdvRtr := convertUint32ToIPv4(key.AdvRouter)

        return fmt.Sprintln("LSA Type:", Type, "LSId:", LSId, "AdvRtr:", AdvRtr)
}

func (server *OSPFServer)dumpAreaStubs() {
        server.logger.Info("=======================Dump Area Stubs======================")
        for key, ent := range server.AreaStubs {
                server.logger.Info("==================================================")
                server.logger.Info(fmt.Sprintln("Vertex Keys:", dumpVertexKey(key)))
                server.logger.Info("==================================================")
                LData := convertUint32ToIPv4(ent.LinkData)
                server.logger.Info(fmt.Sprintln("VertexKeys:", dumpVertexKey(ent.NbrVertexKey), "Cost:", ent.NbrVertexCost, "LinkData:", LData))
                server.logger.Info("==================================================")
                server.logger.Info(fmt.Sprintln("Lsa Key:", dumpLsaKey(ent.LsaKey)))
                server.logger.Info(fmt.Sprintln("AreaId:", ent.AreaId))
                server.logger.Info(fmt.Sprintln("LinkStateId:", ent.LinkStateId))
        }
        server.logger.Info("==================================================")

}

func (server *OSPFServer)dumpAreaGraph() {
        server.logger.Info("=======================Dump Area Graph======================")
        for key, ent := range server.AreaGraph {
                server.logger.Info("==================================================")
                server.logger.Info(fmt.Sprintln("Vertex Keys:", dumpVertexKey(key)))
                server.logger.Info("==================================================")
                if len(ent.NbrVertexKey) != 0 {
                        server.logger.Info("List of Neighbor Vertices(except stub)")
                } else {
                        server.logger.Info("No Neighbor Vertices(except stub)")
                }
                for i := 0; i < len(ent.NbrVertexKey); i++ {
                        LData := convertUint32ToIPv4(ent.LinkData[ent.NbrVertexKey[i]])
                        server.logger.Info(fmt.Sprintln("VertexKeys:", dumpVertexKey(ent.NbrVertexKey[i]), "Cost:", ent.NbrVertexCost[i], "LinkData:", LData))
                }
                server.logger.Info("==================================================")
                server.logger.Info(fmt.Sprintln("Lsa Key:", dumpLsaKey(ent.LsaKey)))
                server.logger.Info(fmt.Sprintln("AreaId:", ent.AreaId))
                server.logger.Info(fmt.Sprintln("Visited:", ent.Visited))
                server.logger.Info(fmt.Sprintln("LinkStateId:", ent.LinkStateId))
        }
        server.logger.Info("==================================================")
}

func (server *OSPFServer)dumpSPFTree() {
        server.logger.Info("=======================Dump SPF Tree======================")
        for key, ent := range server.SPFTree {
                server.logger.Info("==================================================")
                server.logger.Info(fmt.Sprintln("Vertex Keys:", dumpVertexKey(key)))
                server.logger.Info("==================================================")
                server.logger.Info(fmt.Sprintln("Distance:", ent.Distance))
                server.logger.Info(fmt.Sprintln("NumOfPaths:", ent.NumOfPaths))
                for i := 0; i < ent.NumOfPaths; i++ {
                        var paths string
                        paths = fmt.Sprintln("Path[", i, "]")
                        for j := 0; j < len(ent.Paths[i]); j++ {
                                paths = paths + fmt.Sprintln("[", dumpVertexKey(ent.Paths[i][j]), "]")
                        }
                        server.logger.Info(fmt.Sprintln(paths))
                }

        }
}


func (server *OSPFServer)UpdateRoutingTbl(vKey VertexKey) {
        for key, ent := range server.SPFTree {
                if vKey == key {
                        server.logger.Info("It's own vertex")
                        continue
                }
                switch key.Type {
                case RouterVertex:
                        server.UpdateRoutingTblForRouter(key, ent, vKey)
                case SNetworkVertex:
                        server.UpdateRoutingTblForSNetwork(key, ent, vKey)
                case TNetworkVertex:
                        server.UpdateRoutingTblForTNetwork(key, ent, vKey)
                }
        }
}

func (server *OSPFServer) spfCalculation() {
        for {
                msg := <-server.StartCalcSPFCh
                server.logger.Info(fmt.Sprintln("Recevd SPF Calculation Notification for:", msg))
                // Create New Routing table
                // Invalidate Old Routing table
                // Backup Old Routing table
                // TODO: Have Per Area Routing Tbl
                server.OldRoutingTbl = nil
                server.OldRoutingTbl = make(map[RoutingTblKey]RoutingTblEntry)
                server.TempRoutingTbl = nil
                server.TempRoutingTbl = make(map[RoutingTblKey]RoutingTblEntry)
                server.OldRoutingTbl = server.RoutingTbl
                flag := false //TODO:Hack
                for key, _ := range server.AreaConfMap {
                        // Initialize Algorithm's Data Structure
                        server.AreaGraph = make(map[VertexKey]Vertex)
                        server.AreaStubs = make(map[VertexKey]StubVertex)
                        server.SPFTree = make(map[VertexKey]TreeVertex)
                        areaId := convertAreaOrRouterIdUint32(string(key.AreaId))
                        vKey, err := server.CreateAreaGraph(areaId)
                        if err != nil {
                                server.logger.Err(fmt.Sprintln("Error while creating graph for areaId:", areaId))
                                flag = true
                                continue
                        }
                        server.logger.Info("=========================Start before Dijkstra=================")
                        server.dumpAreaGraph()
                        server.dumpAreaStubs()
                        server.logger.Info("=========================End before Dijkstra=================")
                        //server.printRouterLsa()
                        err = server.ExecuteDijkstra(vKey, areaId)
                        if err != nil {
                                server.logger.Err(fmt.Sprintln("Error while executing Dijkstra for areaId:", areaId))
                                flag = true
                                continue
                        }
                        server.logger.Info("=========================Start after Dijkstra=================")
                        server.dumpAreaGraph()
                        server.dumpAreaStubs()
                        server.dumpSPFTree()
                        server.logger.Info("=========================End after Dijkstra=================")
                        server.UpdateRoutingTbl(vKey)
                        server.AreaGraph = nil
                        server.AreaStubs = nil
                        server.SPFTree = nil
                }
                if flag == false {
                        server.InstallRoutingTbl()
                        server.RoutingTbl = nil
                        server.RoutingTbl = make(map[RoutingTblKey]RoutingTblEntry)
                        server.RoutingTbl = server.TempRoutingTbl
                        server.TempRoutingTbl = nil
                        server.OldRoutingTbl = nil
                } else {
                        server.logger.Info("Some Error in Routing Table Generation")
                }
                server.dumpRoutingTbl()
                server.DoneCalcSPFCh <- true
        }
}