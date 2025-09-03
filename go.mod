module github.com/craigbucher/learn-http-servers

go 1.24.4

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
)


// The 'go.sum' file contains cryptographic checksums (hashes) for each version of each dependency your 
// project uses.

// Here's what it does:

// Security: It ensures that the exact same code is downloaded every time, preventing supply chain 
// attacks where someone might replace a package with malicious code.

// Reproducibility: Anyone who clones your project and runs go mod download will get exactly the same 
// versions of dependencies that you used.

// Integrity: If a dependency's code changes (which shouldn't happen for a tagged version), Go will 
// detect this and refuse to use the corrupted version. 