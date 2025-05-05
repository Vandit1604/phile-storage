import { useEffect, useState } from 'react'
import { getAllPeers, discoverFileAcrossPeers } from '../api'

export default function FileBrowser() {
  const [filename, setFilename] = useState('')
  const [peers, setPeers] = useState([])
  const [allPeers, setAllPeers] = useState([])
  const [source, setSource] = useState(null)

  useEffect(() => {
    getAllPeers().then(obj => setAllPeers(Object.values(obj)))
  }, [])

  async function handleDiscover() {
    if (!filename) return
    const result = await discoverFileAcrossPeers(filename, allPeers)
    if (result) {
      setPeers(result.holders)
      setSource(result.source)
    } else {
      setPeers([])
      setSource(null)
    }
  }

  return (
    <div className="flex flex-col items-center space-y-4 text-center">
      <div>
        <h2 className="text-2xl font-semibold text-gray-900">🔍 Discover File</h2>
        <p className="text-sm text-gray-500 mt-1">Find out which peers store a file</p>
      </div>

      <div className="w-full flex flex-col sm:flex-row items-center justify-center gap-2">
        <input
          type="text"
          value={filename}
          onChange={(e) => setFilename(e.target.value)}
          placeholder="filename.txt"
          className="w-full sm:w-auto flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
        />
        <button
          onClick={handleDiscover}
          className="w-full sm:w-auto rounded-lg bg-green-600 px-4 py-2 text-sm font-medium text-white hover:bg-green-700 transition"
        >
          Search
        </button>
      </div>

      {source && (
        <p className="text-xs text-gray-500">
          🔗 Queried via: <span className="font-mono">{source}</span>
        </p>
      )}

      {peers.length > 0 && (
        <div className="w-full">
          <h3 className="text-sm font-medium text-gray-700 mb-2">Peers with this file:</h3>
          <ul className="space-y-2">
            {peers.map((p, i) => (
              <li
                key={i}
                className="flex justify-between items-center rounded-lg border px-4 py-2 text-sm bg-gray-50"
              >
                <span className="truncate">{p}</span>
                <a
                  href={`http://${p}/download?filename=${filename}`}
                  className="text-green-600 hover:underline font-medium"
                >
                  Download
                </a>
              </li>
            ))}
          </ul>
        </div>
      )}

      {peers.length === 0 && filename && (
        <p className="text-sm text-red-500 italic">File not found on any peer.</p>
      )}
    </div>
  )
}

