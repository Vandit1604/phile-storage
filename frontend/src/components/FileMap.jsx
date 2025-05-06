
import { useEffect, useState } from 'react'

export default function FileMap() {
  const [files, setFiles] = useState({})

  useEffect(() => {
    async function fetchFiles() {
      try {
        const res = await fetch('http://127.0.0.1:5001/files')
        const data = await res.json()
        setFiles(data)
      } catch {
        setFiles({})
      }
    }

    fetchFiles() // initial fetch
    const interval = setInterval(fetchFiles, 5000) // auto-refresh every 5s
    return () => clearInterval(interval) // cleanup on unmount
  }, [])

  return (
    <div className="flex flex-col items-center text-center space-y-4">
      <div>
        <h2 className="text-2xl font-semibold text-gray-900">📂 File Peer Map</h2>
        <p className="text-sm text-gray-500 mt-1">Which peers have which files</p>
      </div>

      {Object.keys(files).length === 0 ? (
        <p className="text-sm text-gray-500 italic">No files found</p>
      ) : (
        <div className="w-full space-y-4">
          {Object.entries(files).map(([filename, peers]) => (
            <div
              key={filename}
              className="text-left bg-gray-50 border border-gray-200 rounded-md p-4"
            >
              <div className="font-medium text-gray-800">{filename}</div>
              <ul className="list-disc ml-5 mt-1 text-sm text-gray-700">
                {peers.map((peer, i) => (
                  <li key={i}>{peer}</li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

