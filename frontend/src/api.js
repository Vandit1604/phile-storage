// Fallback peer for peer list discovery
const DEFAULT_PEER = 'http://127.0.0.1:5001'

// Upload a file to a peer. The peer hashes it and returns its CID — the
// content address the file is now reachable by from anywhere in the network.
export async function uploadFile(file, peerAddress) {
  const form = new FormData()
  form.append('file', file)

  const res = await fetch(`http://${peerAddress}/upload`, {
    method: 'POST',
    body: form,
  })

  if (!res.ok) throw new Error(await res.text())
  return await res.json() // { filename, cid }
}

// Get peer list from default peer
export async function getAllPeers() {
  const res = await fetch(`${DEFAULT_PEER}/peers`)
  if (!res.ok) throw new Error(await res.text())
  return await res.json() // { uuid: ip:port, ... }
}

// Resolve a filename on one peer to its CID and current holders.
export async function discoverFile(filename, peerAddress) {
  const res = await fetch(`http://${peerAddress}/discover?filename=${encodeURIComponent(filename)}`)
  if (!res.ok) throw new Error(await res.text())
  return await res.json() // { cid, holders }
}

// Search across all peers. The query is fuzzy-matched against filenames, or
// treated as a direct lookup when it is a CID. Results from every peer are
// merged and de-duplicated by CID (holders unioned).
export async function searchAcrossPeers(query, peerAddresses) {
  const byCid = new Map()

  await Promise.all(
    peerAddresses.map(async (addr) => {
      try {
        const res = await fetch(`http://${addr}/search?q=${encodeURIComponent(query)}`)
        if (!res.ok) return
        const results = await res.json()
        for (const r of results || []) {
          const existing = byCid.get(r.cid)
          const holders = new Set([...(existing?.holders || []), ...(r.holders || [])])
          byCid.set(r.cid, {
            cid: r.cid,
            filename: r.filename || existing?.filename || '',
            holders: [...holders],
          })
        }
      } catch {
        // Skip unreachable peers.
      }
    })
  )

  return [...byCid.values()]
}

// Fetch a peer's libp2p network identity (PeerID + multiaddrs).
export async function getP2PInfo(peerAddress) {
  const res = await fetch(`http://${peerAddress}/p2p/info`)
  if (!res.ok) throw new Error(await res.text())
  return await res.json() // { enabled, peerID, addrs }
}
