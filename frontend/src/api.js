// Fallback peer for peer list discovery
const DEFAULT_PEER = 'http://127.0.0.1:5001'

// Upload file to specific peer
export async function uploadFile(file, peerAddress) {
  const form = new FormData()
  form.append('file', file)

  const res = await fetch(`http://${peerAddress}/upload`, {
    method: 'POST',
    body: form,
  })

  if (!res.ok) throw new Error(await res.text())
  return await res.text()
}

// Get peer list from default peer
export async function getAllPeers() {
  const res = await fetch(`${DEFAULT_PEER}/peers`)
  if (!res.ok) throw new Error(await res.text())
  return await res.json() // { uuid: ip:port, ... }
}

// Discover file on one peer
export async function discoverFile(filename, peerAddress) {
  const res = await fetch(`http://${peerAddress}/discover?filename=${filename}`)
  if (!res.ok) throw new Error(await res.text())
  return await res.json()
}

// Discover file across all peers (first match wins)
export async function discoverFileAcrossPeers(filename, peerAddresses) {
  for (const addr of peerAddresses) {
    try {
      const res = await fetch(`http://${addr}/discover?filename=${filename}`)
      if (res.ok) {
        const peers = await res.json()
        if (peers.length > 0) {
          return { source: addr, holders: peers }
        }
      }
    } catch (err) {
      // Silent failover to next peer
    }
  }
  return null
}

