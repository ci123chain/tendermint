package core

import (
	"github.com/tendermint/tendermint/p2p"
	rpctypes "github.com/tendermint/tendermint/rpc/lib/types"
	"strconv"
	"strings"
)

///  236ba75f825ddfe06471d19b45c683a862e9efe6@test-chain-tls-server.gw001.oneitfarm.com:7443;

func AddPeers(ctx *rpctypes.Context, peers string) error {

	pList, err := parsePeers(peers)
	if err != nil {
		return err
	}

	ourAddr, err := p2pTransport.NodeInfo().NetAddress()
	if err != nil {
		return err
	}
	for _, p := range pList {
		err := addrBook.AddAddress(&p, ourAddr)
		if  err != nil {
			return err
		}
	}
	addrBook.Save()
	return nil
}

func parsePeers(peers string) ([]p2p.NetAddress, error) {
	var peerList = make([]p2p.NetAddress, 0)
	list := strings.Split(peers, ";")
	for _, l := range list {
		var p p2p.NetAddress
		idList := strings.Split(l, "@")
		p.ID = p2p.ID(idList[0])
		ipList := strings.Split(idList[1], ":")
		p.Host = ipList[0]
		port, err := strconv.ParseUint(ipList[1], 10, 64)
		if err != nil {
			return nil, err
		}
		p.Port = uint16(port)
		peerList = append(peerList, p)
	}
	return peerList, nil
}
