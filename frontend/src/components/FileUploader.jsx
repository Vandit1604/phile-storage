import { useEffect, useState } from 'react'
import { getAllPeers, uploadFile } from '../api'

export default function FileUploader() {
  const [file, setFile] = useState(null)
  const [status, setStatus] = useState('')
  const [cid, setCid] = useState('')
  const [allPeers, setAllPeers] = useState([])
  const [selectedPeer, setSelectedPeer] = useState('')

  useEffect(() => {
    getAllPeers()
      .then(obj => {
        const peers = Object.values(obj)
        setAllPeers(peers)
        if (peers.length > 0) setSelectedPeer(peers[0])
      })
      .catch(() => setAllPeers([]))
  }, [])

  async function handleUpload() {
    if (!file || !selectedPeer) return
    try {
      const { filename, cid } = await uploadFile(file, selectedPeer)
      setStatus(`Stored ${filename}`)
      setCid(cid)
    } catch (err) {
      setStatus(`Upload failed: ${err.message}`)
      setCid('')
    }
  }

  return (
    <>
      <p className="text-soft text-sm">Hash a file and store it on a peer.</p>

      <input
        type="file"
        onChange={(e) => setFile(e.target.files[0])}
        className="field cursor-pointer file:mr-3 file:rounded file:border-0 file:bg-[var(--color-lime)] file:px-3 file:py-1 file:font-semibold file:text-[var(--color-forest)]"
      />

      <div className="flex flex-col gap-2 sm:flex-row">
        <select
          value={selectedPeer}
          onChange={(e) => setSelectedPeer(e.target.value)}
          className="field sm:flex-1"
        >
          {allPeers.map((p, i) => (
            <option key={i} value={p} className="text-black">{p}</option>
          ))}
        </select>
        <button onClick={handleUpload} className="btn">Upload</button>
      </div>

      {status && <p className="text-dim text-sm">{status}</p>}

      {cid && (
        <div className="row">
          <div className="eyebrow text-soft">Content ID</div>
          <div className="mono text-[var(--color-lime)] text-sm mt-1">{cid}</div>
        </div>
      )}
    </>
  )
}
