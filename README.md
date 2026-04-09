# UFOs
### Unidentifiable File/Object store

UFOs is a **Zero-Trust, Decentralized, and Social Filesystem** designed for sovereign data ownership. It allows users to store, share, and navigate UFOs across a network of private servers without the host ever seeing the filenames, directory structures, or file contents.

## 🚀 The Core Philosophy

- **Zero Trust**: The server is a "Blind Vault". It stores encrypted BLOBs and HMAC-SHA3-256 search indices. It never sees plaintext filenames, directory trees, or file contents.
- **Hierarchical Privacy**: Directory structures are virtualized. The client germinates the tree using hashed path segments, making your folder structure a secret known only to the user.
- **Social Orbit**: An asynchronous sharing protocol based on X25519 ECDH (Elliptic-Curve Diffie-Hellman). Share UFOs securely with friends across different domains.
- **Titan-Grade Streaming**: Built with Go's `io.Pipe` and `io.TeeReader` primitives, UFOs handles multi-gigabyte files with constant, minimal memory pressure.

## 🛠 Features

- **Hashed Path Germination**: Automatic creation of parent "folder" objects during upload/update to maintain a navigable hierarchy.
- **Multi-User Access Envelopes**: Granular access control where authorized guests receive a specialized cryptographic envelope that wraps the file decryption key.
- **Deterministic Structural Integrity**: Unique constraints on hashed path/name combinations prevent collisions without leaking plaintext metadata.
- **Object Hashing**: On-the-fly SHA3-256 verification during download to detect server-side tampering.
- **Local Vault**: Local storage of private keys and master keys, protected by a passphrase-derived KDF (Argon2id).

## 📐 Architecture

### Cryptographic Stack
- **Identity**: Ed25519 (Signing) and X25519 (Key Exchange).
- **Data Encryption**: AES-256-CTR for high-performance file streaming.
- **Key Wrapping**: AES-256-GCM for metadata BLOBs and data encryption keys.
- **Indexing**: HMAC-SHA3-256 for search tags and path prefixes to ensure "Server Blindness".

### Security Posture
- **Transport**: All API calls are signed using the persona's private signing key.
- **Replay Protection**: Unique request IDs are tracked by the server to prevent replay attacks.
- **Memory Hygiene**: Strict use of `clear()` for raw key buffers and sensitive metadata structs.

---

## 💻 CLI Command Map

| Command | Param Shortcuts | Description |
| :--- | :--- | :--- |
| `init` | - | Initializes the local vault. |
| `new` | `n` | Bootstrap a new persona with a remote UFOs server. |
| `register` | `n, t` | Register a persona with a remote UFOs server using the registration token from 'new' command, or the token the server admin sets in the 'UFO_BOOTSTRAP_TOKEN' env variable (for initial user bootstrapping). |
| `ping` | `d` | Checks to see if a given server is responsive. |
| `upload` | `n, f, p, t, a` | Upload a UFO, generate path hierarchy, and set access. |
| `update` | `n, i, f, p, t, a` | Modify name, path, tags, or access list for a given UFO. |
| `download` | `n, i, h, t` | Download/stream from a UFOs server. |
| `list` | `n, p` | List the UFO hierarchy for a specific prefix. |
| `search` | `n, p, t` | Find UFOs globally using one or more hashed tags. |
| `info` | `n, i` | View detailed metadata and access lists for a specific UFO. |
| `remove` | `n, i` | Remove a UFO or recursively remove a directory tree. |
| `orbit add` | - | Add a satellite (a fully qualified persona - `id@domain`) to your social orbit. |
| `orbit list` | - | View all satellites currently in your orbit. |
| `orbit remove` | - | Remove a satellite from your orbit. |

---

## 📡 Server API Map

### Server health
- `GET /healthz`: Check server status

### Authentication & Registration
- `POST /api/init`: Bootstrap a new identity (persona)
- `POST /api/personas`: Register a new persona.
- `GET /api/personas/{id}`: Public key discovery for social handshakes.

### UFO Management
- `POST /api/ufos`: Register UFO metadata (File or Folder).
- `PUT /api/ufos/{uuid}`: Ingest raw binary object bytes.
- `PATCH /api/ufos/{uuid}`: Atomic update of metadata and indices.
- `GET /api/ufos/{uuid}`: Retrieve object bytes and extraction headers.
- `HEAD /api/ufos/{uuid}`: Peek at UFO metadata headers.
- `DELETE /api/ufos/{uuid}`: Permanent removal of a UFO and its indices.

### Discovery & Social
- `GET /api/ufos`: List UFOs at a given `prefix` or with provided `tags` (hashed)
- `GET /api/orbit`: Retrieve authorized satellites for the active persona.
- `POST /api/orbit`: Add new satellite for the active persona's orbit.
- `GET /api/orbit/{id}`: Retrieve a specific satellite's public keys and metadata.
- `DELETE /api/orbit/{id}`: Remove a satellite from the active persona's orbit.

---

## 🏗 Roadmap
- [ ] Implement LetsEncrypt cert protocol to enable https

Future Ideas:
- TUI (Terminal User Interface) for visual navigation?
- Mounting: Mount your UFOs as a local encrypted drive?