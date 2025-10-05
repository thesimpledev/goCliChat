# goChat

A self-hosted, encrypted chat server and client written in Go. goChat uses public-key cryptography to provide password-free authentication and encrypted messaging over TCP.

## Architecture

### Security Model

goChat uses a Trust On First Use (TOFU) security model, similar to SSH:

- Users are responsible for verifying they're connecting to the correct server
- No centralized certificate authority or PKI
- Key-based authentication eliminates password storage and transmission risks
- Each user maintains their own key pairs locally

### Cryptographic Design

The system uses the Curve25519 family of algorithms:

**Ed25519** for message signing
- Signature size: 64 bytes
- Public key size: 32 bytes
- Used to authenticate message senders

**X25519** for encryption
- Public key size: 32 bytes
- Encryption overhead: 16 bytes (Poly1305 MAC with ChaCha20-Poly1305)
- Provides confidentiality using ECDH shared secrets

**Message Flow:**
1. Client signs message with their Ed25519 private key
2. Client encrypts (message + signature) with server's X25519 public key
3. Server decrypts with its X25519 private key
4. Server verifies signature using client's Ed25519 public key

This provides both authentication (you know who sent it) and confidentiality (only the intended recipient can read it).

### Protocol Specification

All communication uses fixed 1024-byte TCP frames for deterministic behavior:

**Wire Format (1024 bytes total):**
- Command: 1 byte (plaintext)
- Encrypted Payload: 1007 bytes
- Poly1305 Authentication Tag: 16 bytes (included in encrypted portion)

**Decrypted Payload (1007 bytes):**
- Username: 32 bytes
- Signature: 64 bytes (Ed25519)
- Message: 911 bytes

Commands are transmitted in plaintext to solve the bootstrap problem (you can't encrypt key exchange requests before you have keys) and to simplify server routing. Command bytes leak no sensitive information.

### Database

The server uses SQLite to persist user information and public keys.

### Network Protocol

**Connection Flow:**

1. Client requests server's public key (CMD_KEY_REQ)
2. Server responds with its X25519 and Ed25519 public keys
3. Client can then either:
   - Register (CMD_REGISTER): Send username + client public keys
   - Connect (CMD_CONNECTION): Establish session with existing username

**Registration:**
- Client generates Ed25519 and X25519 key pairs
- Client sends CMD_REGISTER with username and both public keys
- Server verifies username is available
- Server stores user information in database
- Server responds with confirmation

**Chat Messages:**
- All messages signed with sender's Ed25519 private key
- All messages encrypted with server's X25519 public key
- Server decrypts, verifies signature, and broadcasts to other users
- Messages re-encrypted for each recipient using their public keys

### Key Rotation

Users can rotate their keys at any time using CMD_ROTATE_KEY:
- Generate new key pairs locally
- Send new public keys to server
- Server updates database
- Old keys are invalidated

This limits damage from key compromise.

## Usage

### Running the Server

Start the server on the default port (8080):

    ./gochat-server

Start the server on a custom port:

    ./gochat-server -port 9000

The server will:
- Create a SQLite database if it doesn't exist
- Generate server key pairs if needed
- Listen for incoming client connections
- Log all activity to stdout

### Running the Client

Connect to a server:

    ./gochat-client username@hostname:port

Register a new user:

    ./gochat-client username@hostname:port -r

The client will:
- Generate key pairs on first run (if registering)
- Request server's public key
- Establish encrypted connection
- Provide interactive chat interface

### Client Commands

While connected, the client accepts standard chat messages. Future versions will support:
- /rotate - Rotate your encryption keys
- /delete - Delete your account
- /quit - Disconnect from server

## Security Considerations

### Trust Model

goChat uses a self-hosted security model similar to SSH:

- **Server Trust**: Users must verify they're connecting to the legitimate server
- **No Central Authority**: There is no PKI or certificate system
- **Key Compromise**: If a private key is compromised, that identity is compromised. Users should rotate keys regularly.

### Known Limitations

- No protection against MITM during initial key exchange
- Traffic analysis possible (message timing and sizes visible)

## Development

### Project Structure

    .
    ├── cmd/
    │   ├── client/       # Client application
    │   └── server/       # Server application
    ├── internal/
    │   ├── db/           # Database layer
    │   └── protocol/     # Protocol constants and frame handling
    ├── go.mod
    ├── LICENSE.md
    └── README.md

### Contributing

Feel free to fork and experiment with the codebase.

### TODO

**Server:**
- [ ] Implement port configuration flag
- [ ] Add graceful shutdown with context cancellation
- [ ] Implement key rotation handler
- [ ] Implement account deletion handler
- [ ] Add database migrations system
- [ ] Implement message persistence
- [ ] Add connection rate limiting

**Client:**
- [ ] Implement key generation and storage
- [ ] Add key rotation command
- [ ] Implement account deletion command
- [ ] Add message history
- [ ] Improve error handling and reconnection
- [ ] Add configuration file support

**Protocol:**
- [ ] Implement frame packing/unpacking
- [ ] Add encryption/decryption layer
- [ ] Implement signature generation and verification
- [ ] Add key serialization

## License

MIT License - see LICENSE.md for details

Copyright © 2025 Steven Stanton

## Acknowledgments

Built with Go using the SSH security model as inspiration.
