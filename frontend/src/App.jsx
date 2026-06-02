import FileUploader from './components/FileUploader'
import FileBrowser from './components/FileBrowser'
import PeerList from './components/PeerList'
import FileMap from './components/FileMap'
import StarBurst from './components/StarBurst'

export default function App() {
  return (
    <div className="min-h-screen p-4 sm:p-8">
      <div className="panel">
        <Header />

        <div className="grid gap-4 mt-10 lg:grid-cols-2">
          <Card index="01" lead="Add" accent="files" slug="upload">
            <FileUploader />
          </Card>
          <Card index="02" lead="Find" accent="anything" slug="discover">
            <FileBrowser />
          </Card>
          <Card index="03" lead="Content" accent="map" slug="filemap">
            <FileMap />
          </Card>
          <Card index="04" lead="Active" accent="peers" slug="peers">
            <PeerList />
          </Card>
        </div>
      </div>
    </div>
  )
}

function Header() {
  return (
    <header className="grid gap-8 md:grid-cols-[auto_1fr] md:items-start">
      <div className="flex items-center gap-4">
        <p className="eyebrow">(00) Phile Storage</p>
      </div>

      <div className="max-w-3xl">
        <h1 className="display">
          Distributed storage for <em>content you own</em>
        </h1>
        <div className="flex items-start gap-5 mt-6">
          <StarBurst className="w-10 h-10 shrink-0 text-[var(--color-forest)]" />
          <p className="text-[var(--color-forest)] max-w-xl leading-relaxed">
            Files are addressed by the hash of their bytes and verified on every
            fetch. Each node carries its own libp2p identity — no central server
            required.
          </p>
        </div>
      </div>
    </header>
  )
}

function Card({ index, lead, accent, slug, children }) {
  return (
    <section aria-labelledby={`${slug}-heading`} className="card">
      <StarBurst className="card__mark" />
      <div className="card__head">
        <div>
          <p className="eyebrow text-[var(--color-lime)]">({index})</p>
          <h2 id={`${slug}-heading`} className="card-title">
            {lead} <em>{accent}</em>
          </h2>
        </div>
        <span className="arrow-circle" aria-hidden="true">↗</span>
      </div>
      <div className="card__body">{children}</div>
    </section>
  )
}
