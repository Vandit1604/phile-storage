import { useEffect, useState } from 'react'
import { getAllPeers, searchAcrossPeers } from '../api'

export default function FileBrowser() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState([])
  const [allPeers, setAllPeers] = useState([])
  const [searched, setSearched] = useState(false)

  useEffect(() => {
    getAllPeers().then(obj => setAllPeers(Object.values(obj))).catch(() => setAllPeers([]))
  }, [])

  async function handleSearch() {
    if (!query.trim()) return
    const found = await searchAcrossPeers(query.trim(), allPeers)
    setResults(found)
    setSearched(true)
  }

  return (
    <>
      <p className="text-soft text-sm">Fuzzy filename match, or paste a content ID.</p>

      <div className="flex flex-col gap-2 sm:flex-row">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
          placeholder="facebook   ·   or   bafkrei…"
          className="field sm:flex-1"
        />
        <button onClick={handleSearch} className="btn">Search</button>
      </div>

      {results.length > 0 && (
        <ul className="flex flex-col gap-2">
          {results.map((r) => (
            <li key={r.cid} className="row">
              <div className="flex items-center justify-between gap-3">
                <span className="font-semibold text-[var(--color-lime)]">{r.filename || '(unnamed)'}</span>
                {r.holders[0] && (
                  <a
                    href={`http://${r.holders[0]}/download?cid=${r.cid}&filename=${encodeURIComponent(r.filename || r.cid)}`}
                    className="text-soft font-semibold text-sm hover:underline shrink-0"
                  >
                    Download
                  </a>
                )}
              </div>
              <div className="mono text-soft text-xs mt-1">{r.cid}</div>
              <div className="mono text-dim text-[11px] mt-0.5">
                {r.holders.length ? r.holders.join(', ') : 'no known holders'}
              </div>
            </li>
          ))}
        </ul>
      )}

      {searched && results.length === 0 && (
        <p className="text-dim text-sm italic">No matches.</p>
      )}
    </>
  )
}
