# Connector session lifecycle (SDK v1)

Connector process lifetime and external transport lifetime are separate. The
engine may keep one plugin process alive while each resource manages device
sessions according to actual demand.

`ConnectorResource.SessionPolicy` accepts these standard values:

| Value | Meaning | Typical connectors |
| --- | --- | --- |
| `scheduled_reuse` | Lazily open a session and retain it only across near-term scheduled demand | High/low frequency Siemens S7 |
| `per_operation` | Open one session for an operation and close before returning | Strict S7 mode, FANUC FOCAS |
| `persistent` | Keep a reconnecting transport while the resource is active | MQTT |
| `pooled` | Use a bounded client/connection pool | SQL databases, Redis |
| `stateless` | Do not expose a connector-managed session | HTTP |
| `idle_timeout` | Reuse a session for a generic configured idle window | Other expensive transports |

`scheduled_reuse` is an additive v1 constant. It does not add a DTO field,
interface method, RPC method, or wire type, so SDK v1 compatibility remains
intact. Protocol-specific tuning belongs in connector connection JSON rather
than the common contract.

## Demand-managed industrial connectors

For `scheduled_reuse`, `CreateResource` validates and stores configuration but
does not connect. One operation obtains at most one device session and uses it
for all chunks in that batch. A healthy session may then serve later batches
when the connector knows another demand is near. It must close on resource stop,
session/protocol failure, configuration replacement, expired maximum age, or no
near-term demand.

The provider serializes one device resource while allowing different resources
to run concurrently under a package-level bound. A high-frequency scheduler
must use latest-only delivery: it skips expired ticks rather than building an
unbounded queue. Read failures naturally retry on a later scan; writes are not
transparently retried.

An idle demand-managed resource is not disconnected. Health normally comes
from real operations. Periodic active probes should be disabled because they
would create artificial demand.

## Capability discovery

Descriptors advertise supported policies with capability strings such as
`session:scheduled_reuse`, `session:per_operation`, `session:persistent`, and
`session:pooled`. The resource's explicit policy remains authoritative. A
connector may normalize documented legacy values without rewriting persisted
data automatically.
