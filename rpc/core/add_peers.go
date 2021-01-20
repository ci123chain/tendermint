package core

import (
	"errors"
	"fmt"
	"github.com/tendermint/tendermint/p2p"
	core_types "github.com/tendermint/tendermint/rpc/core/types"
	rpctypes "github.com/tendermint/tendermint/rpc/lib/types"
	"net"
	"strconv"
	"strings"
)

///  236ba75f825ddfe06471d19b45c683a862e9efe6@test-chain-tls-server.gw001.oneitfarm.com:127.0.0.1:7443;

func AddPeers(ctx *rpctypes.Context, peers string) (*core_types.ResultAddPeers,error) {

	pList, err := parsePeers(peers)
	if err != nil {
		return &core_types.ResultAddPeers{},err
	}

	ourAddr, err := p2pTransport.NodeInfo().NetAddress()
	if err != nil {
		return &core_types.ResultAddPeers{},err
	}
	for _, p := range pList {
		err := addrBook.AddAddress(&p, ourAddr)
		if  err != nil {
			return &core_types.ResultAddPeers{},err
		}
	}
	addrBook.Save()
	return &core_types.ResultAddPeers{},nil
}

func parsePeers(peers string) ([]p2p.NetAddress, error) {
	var peerList = make([]p2p.NetAddress, 0)
	list := strings.Split(peers, ";")
	for _, l := range list {
		var p p2p.NetAddress
		idList := strings.Split(l, "@")
		if len(idList) != 2 {
			return nil, errors.New(fmt.Sprintf("invalid peers: %s", peers))
		}
		p.ID = p2p.ID(idList[0])
		ipList := strings.Split(idList[1], ":")
		if len(ipList) < 2 || len(ipList) > 3 {
			return nil, errors.New(fmt.Sprintf("invalid peers: %s", peers))
		}
		p.Host = ipList[0]
		var port uint64
		var err error
		if len(ipList) == 2 {
			port, err = strconv.ParseUint(ipList[1], 10, 64)
			if err != nil {
				return nil, err
			}
		}else {
			port, err = strconv.ParseUint(ipList[2], 10, 64)
			if err != nil {
				return nil, err
			}
			p.IP = net.ParseIP(ipList[1])
		}

		p.Port = uint16(port)
		peerList = append(peerList, p)
	}
	return peerList, nil
}
