import { useEffect, useState } from 'react'
import { getAllPeers } from '../api'

export default function PeerList() {
  const [peers, setPeers] = useState({})

  useEffect(() => {
    getAllPeers().then(setPeers).catch(() => setPeers({}))
  }, [])

  return (
    <div className="flex flex-col items-center text-center space-y-4">
      <div>
        <h2 className="text-2xl font-semibold text-gray-900">🌐 Active Peers</h2>
        <p className="text-sm text-gray-500 mt-1">List of all currently registered peer nodes</p>
      </div>

      <div className="w-full space-y-2">
        {Object.keys(peers).length === 0 ? (
          <p className="text-sm text-gray-500 italic">No peers registered</p>
        ) : (
          <ul className="space-y-2">
            {Object.entries(peers).map(([uuid, addr]) => (
              <li
                key={uuid}
                className="flex flex-col sm:flex-row sm:justify-between sm:items-center border rounded-md px-4 py-3 bg-gray-50 text-sm text-gray-800"
              >
                <div className="truncate font-mono text-xs text-gray-600">{uuid}</div>
                <div className="mt-1 sm:mt-0 font-medium text-gray-900">{addr}</div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}

