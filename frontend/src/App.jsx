
import FileUploader from './components/FileUploader'
import FileBrowser from './components/FileBrowser'
import PeerList from './components/PeerList'
import FileMap from './components/FileMap'

export default function App() {
  return (
    <div className="min-h-screen bg-gray-50 flex flex-col items-center px-4 py-12">
      <div className="w-full max-w-5xl space-y-10">
        <header className="text-center">
          <h1 className="text-5xl font-bold tracking-tight text-gray-900">Phile Storage</h1>
          <p className="text-gray-500 mt-2 text-sm">Peer-to-peer file sharing dashboard</p>
        </header>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Card>
            <FileUploader />
          </Card>

          <Card>
            <FileBrowser />
          </Card>
        </div>

        <Card>
          <FileMap />
        </Card>

        <Card>
          <PeerList />
        </Card>
      </div>
    </div>
  )
}

function Card({ children }) {
  return (
    <div className="bg-white border border-gray-200 rounded-2xl shadow-sm hover:shadow-md transition-shadow duration-200 p-6">
      {children}
    </div>
  )
}

