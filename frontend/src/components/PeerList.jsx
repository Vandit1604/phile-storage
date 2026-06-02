import { useEffect, useState } from 'react'
import { getAllPeers, getP2PInfo } from '../api'

export default function PeerList() {
  const [peers, setPeers] = useState({})
  const [identities, setIdentities] = useState({}) // addr -> peerID

  useEffect(() => {
    async function fetchPeers() {
      try {
        const data = await getAllPeers()
        setPeers(data)

        const ids = {}
        await Promise.all(
          Object.values(data).map(async (addr) => {
            try {
              const info = await getP2PInfo(addr)
              if (info.enabled) ids[addr] = info.peerID
            } catch {
              // Peer may not have libp2p enabled; skip.
            }
          })
        )
        setIdentities(ids)
      } catch {
        setPeers({})
        setIdentities({})
      }
    }
    fetchPeers()
    const interval = setInterval(fetchPeers, 5000)
    return () => clearInterval(interval)
  }, [])

  const entries = Object.entries(peers)

  return (
    <>
      <p className="text-soft text-sm">Live nodes and their libp2p identities.</p>

      {entries.length === 0 ? (
        <p className="text-dim text-sm italic">No peers registered.</p>
      ) : (
        <ul className="flex flex-col gap-2">
          {entries.map(([uuid, addr]) => (
            <li key={uuid} className="row">
              <div className="flex items-center justify-between gap-3">
                <span className="mono text-dim text-xs">{uuid}</span>
                <span className="font-semibold text-[var(--color-lime)]">{addr}</span>
              </div>
              {identities[addr] && (
                <div className="mono text-soft text-[11px] mt-1">PeerID: {identities[addr]}</div>
              )}
            </li>
          ))}
        </ul>
      )}
    </>
  )
}
