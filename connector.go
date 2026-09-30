package sdk

import (
	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
	transport "github.com/punk-one/iot-rule-plugin-sdk/internal/transport"
)

type ConnectorPlugin = core.ConnectorPlugin

type ConnectorRPC = transport.ConnectorRPC
type ConnectorRPCServer = transport.ConnectorRPCServer
type ConnectorRPCClient = transport.ConnectorRPCClient
