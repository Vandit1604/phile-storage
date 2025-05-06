# 🗂️ Phile Storage

![Screenshot From 2025-05-07 00-10-04](https://github.com/user-attachments/assets/d7fc1ecf-5a16-4bfe-874d-39699faaafa7)


**Phile Storage** is a simple peer-to-peer distributed file sharing system built with Go and React. It enables decentralized file uploads, discovery, and downloads across multiple dynamically spawned peers — with a real-time dashboard.

---

## 🔧 Features

### ✅ Backend (Go + Redis + etcd)
- 📦 **Peer Registration** with `etcd`
  - Each peer registers itself and heartbeats to stay active
- 📤 **File Upload**
  - Upload files to any peer node; stored locally under `data/<peer-uuid>/`
  - Metadata stored in Redis
- 🔍 **File Discovery**
  - Discover which peers hold a specific file
- 📥 **File Download**
  - Download from local peer if available
  - Else pull from another peer and cache locally
- 📄 **File Map Endpoint**
  - `/files` lists all known files and the peers that store them
- 🌐 **Peer Map Endpoint**
  - `/peers` lists all active peer UUID → address mappings

---

### 🖥️ Frontend (React + Tailwind)
- 📤 **File Uploader**
  - Select a peer from dropdown and upload a file
- 🔍 **File Browser**
  - Search for a filename and see which peers have it
  - Direct download links to those peers
- 🌐 **Peer List**
  - Live list of currently connected peer nodes
- 🗃️ **File-Peer Map**
  - See all files in the system and which peers store them
- 🔁 **Auto-refresh**
  - Peer and file maps update every 5 seconds

---

## 🚀 Getting Started

### Prerequisites
- Go 1.20+
- Node 18+ / npm
- Docker (for Redis + etcd)

### Setup

```bash
# Start Redis and etcd
make start-docker

# Build and run 3 peer nodes
make run-peers PEER_COUNT=3

# Start frontend
cd frontend
npm install
npm run dev
````

Open browser: [http://localhost:5173](http://localhost:5173)

## 🧩 Tech Stack

* **Go** (net/http, etcd client, Redis)
* **Redis** for metadata storage
* **etcd** for peer discovery & TTL-based registry
* **React + Tailwind CSS** for dashboard
* **Makefile** for streamlined development

