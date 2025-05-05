import { useEffect, useState } from 'react'
import { getAllPeers, uploadFile } from '../api'

export default function FileUploader() {
  const [file, setFile] = useState(null)
  const [status, setStatus] = useState('')
  const [allPeers, setAllPeers] = useState([])
  const [selectedPeer, setSelectedPeer] = useState('')

  useEffect(() => {
    getAllPeers().then(obj => {
      const peers = Object.values(obj)
      setAllPeers(peers)
      if (peers.length > 0) setSelectedPeer(peers[0])
    })
  }, [])

  async function handleUpload() {
    if (!file || !selectedPeer) return
    try {
      const msg = await uploadFile(file, selectedPeer)
      setStatus(msg)
    } catch (err) {
      setStatus(`❌ ${err.message}`)
    }
  }

  return (
    <div className="flex flex-col items-center text-center space-y-4">
      <div>
        <h2 className="text-2xl font-semibold text-gray-900">📤 Upload File</h2>
        <p className="text-sm text-gray-500 mt-1">Send a file to a selected peer</p>
      </div>

      <input
        type="file"
        onChange={(e) => setFile(e.target.files[0])}
        className="block w-full text-sm text-gray-700 file:mr-4 file:py-2 file:px-4
                 file:rounded-lg file:border-0 file:text-sm file:font-semibold
                 file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 cursor-pointer"
      />

      <select
        value={selectedPeer}
        onChange={(e) => setSelectedPeer(e.target.value)}
        className="w-full sm:w-auto rounded-lg border border-gray-300 px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        {allPeers.map((peer, i) => (
          <option key={i} value={peer}>
            {peer}
          </option>
        ))}
      </select>

      <button
        onClick={handleUpload}
        className="w-full sm:w-auto bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium px-6 py-2 rounded-lg transition"
      >
        Upload
      </button>

      {status && (
        <p className="text-sm text-gray-700 bg-gray-50 border border-gray-200 rounded-md px-4 py-2 w-full text-center">
          {status}
        </p>
      )}
    </div>
  )
}

