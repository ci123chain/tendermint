package commands

import (
	"fmt"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/p2p/tls"

	"github.com/spf13/cobra"

	cmn "github.com/tendermint/tendermint/libs/common"
	nm "github.com/tendermint/tendermint/node"
)

// AddNodeFlags exposes some common configuration options on the command-line
// These are exposed for convenience of commands embedding a tendermint node
func AddNodeFlags(cmd *cobra.Command) {
	// bind flags
	cmd.Flags().String("moniker", config.Moniker, "Node Name")

	// priv val flags
	cmd.Flags().String("priv_validator_laddr", config.PrivValidatorListenAddr, "Socket address to listen on for connections from external priv_validator process")

	// node flags
	cmd.Flags().Bool("fast_sync", config.FastSyncMode, "Fast blockchain syncing")

	// abci flags
	cmd.Flags().String("proxy_app", config.ProxyApp, "Proxy app address, or one of: 'kvstore', 'persistent_kvstore', 'counter', 'counter_serial' or 'noop' for local testing.")
	cmd.Flags().String("abci", config.ABCI, "Specify abci transport (socket | grpc)")

	// rpc flags
	cmd.Flags().String("rpc.laddr", config.RPC.ListenAddress, "RPC listen address. Port required")
	cmd.Flags().String("rpc.grpc_laddr", config.RPC.GRPCListenAddress, "GRPC listen address (BroadcastTx only). Port required")
	cmd.Flags().Bool("rpc.unsafe", config.RPC.Unsafe, "Enabled unsafe rpc methods")

	// p2p flags
	cmd.Flags().String("p2p.laddr", config.P2P.ListenAddress, "Node listen address. (0.0.0.0:0 means any interface, any port)")
	cmd.Flags().String("p2p.seeds", config.P2P.Seeds, "Comma-delimited ID@host:port seed nodes")
	cmd.Flags().String("p2p.persistent_peers", config.P2P.PersistentPeers, "Comma-delimited ID@host:port persistent peers")
	cmd.Flags().Bool("p2p.upnp", config.P2P.UPNP, "Enable/disable UPNP port forwarding")
	cmd.Flags().Bool("p2p.pex", config.P2P.PexReactor, "Enable/disable Peer-Exchange")
	cmd.Flags().Bool("p2p.seed_mode", config.P2P.SeedMode, "Enable/disable seed mode")
	cmd.Flags().String("p2p.private_peer_ids", config.P2P.PrivatePeerIDs, "Comma-delimited private peer IDs")

	//tls config
	cmd.Flags().Bool("p2p.tls_option", config.P2P.TLSOption, "Enable/disable tls_option")
	cmd.Flags().Int("tls_config.bind_address_port", config.TLSConfig.BindAddressPort, "tls listen port")
	cmd.Flags().String("tls_config.remote_address_host", config.TLSConfig.RemoteAddressHOST, "remote host")
	cmd.Flags().Int("tls_config.remote_address_port", config.TLSConfig.RemoteAddressPort, "remote port")
	cmd.Flags().Bool("tls_config.remote_tls_insecure_skip_verify", config.TLSConfig.RemoteTLSInsecureSkipVerify, "Enable/Disable gateway skip verify")

	// consensus flags
	cmd.Flags().Bool("consensus.create_empty_blocks", config.Consensus.CreateEmptyBlocks, "Set this to false to only produce blocks when there are txs or when the AppHash changes")
}

// NewRunNodeCmd returns the command that allows the CLI to start a node.
// It can be used with a custom PrivValidator and in-process ABCI application.
func NewRunNodeCmd(nodeProvider nm.NodeProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Run the tendermint node",
		RunE: func(cmd *cobra.Command, args []string) error {
			var b = config
			fmt.Println(b.P2P.PersistentPeers)
			n, err := nodeProvider(config, logger)
			if err != nil {
				return fmt.Errorf("Failed to create node: %v", err)
			}

			// Stop upon receiving SIGTERM or CTRL-C.
			cmn.TrapSignal(logger, func() {
				if n.IsRunning() {
					n.Stop()
				}
			})

			var newCfg = n.Config()
			//var node_store = p2p.NewNodeStore()
			p2p.SetP2PConfig(newCfg)
			p2p.SetLogger(logger)
			//p2p.SetStore(node_store)
			tls.SetTLSConfig(newCfg.TLSConfig)
			tls.SetLogger(logger)
			if newCfg.P2P.TLSOption {
				tls.NewTLS()
				logger.Info("Started TLS")
			}

			if err := n.Start(); err != nil {
				return fmt.Errorf("Failed to start node: %v", err)
			}
			logger.Info("Started node", "nodeInfo", n.Switch().NodeInfo())
			//var newCfg = n.Config()
			//p2p.SetP2PConfig(newCfg.P2P)
			//p2p.SetLogger(logger)
			//tls.SetTLSConfig(newCfg.P2P.TLSConfig)
			//tls.SetLogger(logger)
			//if newCfg.P2P.TLSOption {
			//	tls.NewTLS()
			//}
			//logger.Info("Started TLS")

			// Run forever.
			select {}
		},
	}

	AddNodeFlags(cmd)
	return cmd
}
