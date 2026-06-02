import { useEffect, useState } from 'react'

export default function FileMap() {
  const [files, setFiles] = useState({})

  useEffect(() => {
    async function fetchFiles() {
      try {
        const res = await fetch('http://127.0.0.1:5001/files')
        setFiles(await res.json())
      } catch {
        setFiles({})
      }
    }
    fetchFiles()
    const interval = setInterval(fetchFiles, 5000)
    return () => clearInterval(interval)
  }, [])

  const entries = Object.entries(files)

  return (
    <>
      <p className="text-soft text-sm">Files this peer knows about.</p>

      {entries.length === 0 ? (
        <p className="text-dim text-sm italic">No files yet.</p>
      ) : (
        <ul className="flex flex-col gap-2">
          {entries.map(([filename, entry]) => (
            <li key={filename} className="row">
              <div className="font-semibold text-[var(--color-lime)]">{filename}</div>
              <div className="mono text-soft text-xs mt-1">{entry.cid}</div>
              <div className="mono text-dim text-[11px] mt-0.5">{(entry.holders || []).join(', ')}</div>
            </li>
          ))}
        </ul>
      )}
    </>
  )
}
