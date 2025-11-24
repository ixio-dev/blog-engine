package deploy

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
	"strings"

	"tech-blog/config"
	"tech-blog/build"
)

type Deployer struct {
	config *config.Config
}

func New(config *config.Config) *Deployer {
	return &Deployer{
		config: config,
	}
}

func (d *Deployer) Deploy() error {
	// Validate configuration
	if d.config.Deploy.Host == "" {
		return fmt.Errorf("deploy.host is required")
	}
	if d.config.Deploy.Path == "" {
		return fmt.Errorf("deploy.path is required")
	}

	// Build the site
	builder, err := build.New(d.config)
	if err != nil {
		return fmt.Errorf("failed to create builder: %w", err)
	}

	if err := builder.Build(d.config.Build.Drafts, false); err != nil {
		return fmt.Errorf("failed to build site: %w", err)
	}

	// Calculate build hash
	hash, err := d.calculateBuildHash()
	if err != nil {
		return fmt.Errorf("failed to calculate build hash: %w", err)
	}

	// Check if deployment is needed
	if d.isAlreadyDeployed(hash) {
		fmt.Println("Build unchanged, skipping deployment")
		return nil
	}

	// Deploy via SCP
	if err := d.deployViaSCP(); err != nil {
		return fmt.Errorf("failed to deploy: %w", err)
	}

	// Save deployment hash
	if err := d.saveDeploymentHash(hash); err != nil {
		return fmt.Errorf("failed to save deployment hash: %w", err)
	}

	fmt.Println("Deployment completed successfully")
	return nil
}

func (d *Deployer) calculateBuildHash() (string, error) {
	hash := sha256.New()

	err := filepath.Walk(d.config.Build.OutputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(hash, file); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (d *Deployer) isAlreadyDeployed(hash string) bool {
	hashFile := ".deployed_hash"
	data, err := os.ReadFile(hashFile)
	if err != nil {
		return false
	}
	return string(data) == hash
}

func (d *Deployer) saveDeploymentHash(hash string) error {
	return os.WriteFile(".deployed_hash", []byte(hash), 0644)
}

func (d *Deployer) deployViaSCP() error {
	// Parse host and user from host string (format: user@host)
	hostParts := strings.Split(d.config.Deploy.Host, "@")
	if len(hostParts) != 2 {
		return fmt.Errorf("invalid host format, expected user@host")
	}
	user := hostParts[0]
	host := hostParts[1]

	// Create SSH client configuration
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			// Try to use SSH agent first, then fallback to key file
			ssh.PublicKeysCallback(d.getSigners),
			ssh.PasswordCallback(d.getPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production, use proper host key checking
	}

	// Connect to SSH server
	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server: %w", err)
	}
	defer client.Close()

	// Create session
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Create remote directory if it doesn't exist
	mkdirCmd := fmt.Sprintf("mkdir -p %s", d.config.Deploy.Path)
	if err := session.Run(mkdirCmd); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Upload files recursively
	return d.uploadDirectory(client, d.config.Build.OutputDir, d.config.Deploy.Path)
}

// uploadDirectory uploads a local directory to a remote directory via SCP
func (d *Deployer) uploadDirectory(client *ssh.Client, localDir, remoteDir string) error {
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from local directory
		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		remotePath := filepath.Join(remoteDir, relPath)
		remotePath = filepath.ToSlash(remotePath) // Convert to forward slashes for Unix paths

		if info.IsDir() {
			// Create remote directory
			session, err := client.NewSession()
			if err != nil {
				return err
			}
			defer session.Close()

			mkdirCmd := fmt.Sprintf("mkdir -p %s", remotePath)
			if err := session.Run(mkdirCmd); err != nil {
				return err
			}

			return nil
		}

		// Upload file
		return d.uploadFile(client, path, remotePath)
	})
}

// uploadFile uploads a single file via SCP protocol
func (d *Deployer) uploadFile(client *ssh.Client, localPath, remotePath string) error {
	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	// Get file info for permissions
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info for %s: %w", localPath, err)
	}

	// Create SSH session for SCP command
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Prepare SCP command
	scpCmd := fmt.Sprintf("scp -t %s", filepath.Dir(remotePath))
	scpStdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	scpStdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start SCP command
	if err := session.Start(scpCmd); err != nil {
		return fmt.Errorf("failed to start SCP command: %w", err)
	}

	// Read response from SCP
	response, err := bufio.NewReader(scpStdout).ReadString('\x00')
	if err != nil || response != "\x00" {
		return fmt.Errorf("SCP command failed to start: %v", response)
	}

	// Send file info to SCP
	fileSize := fileInfo.Size()
	fileMode := fileInfo.Mode()
	fmt.Fprintf(scpStdin, "C%04o %d %s\n", fileMode.Perm(), fileSize, filepath.Base(remotePath))

	// Read response after sending file info
	response, err = bufio.NewReader(scpStdout).ReadString('\x00')
	if err != nil || response != "\x00" {
		return fmt.Errorf("SCP command failed after sending file info: %v", response)
	}

	// Copy file contents
	if _, err := io.Copy(scpStdin, localFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Send end-of-file marker
	fmt.Fprintf(scpStdin, "\x00")

	// Read final response
	response, err = bufio.NewReader(scpStdout).ReadString('\x00')
	if err != nil || response != "\x00" {
		return fmt.Errorf("SCP command failed after sending file: %v", response)
	}

	// Close stdin to finish the transfer
	scpStdin.Close()

	// Wait for session to complete
	return session.Wait()
}

// getSigners returns SSH key signers for authentication
func (d *Deployer) getSigners() ([]ssh.Signer, error) {
	// Try standard SSH key paths
	paths := []string{
		os.Getenv("HOME") + "/.ssh/id_rsa",   // Standard RSA key
		os.Getenv("HOME") + "/.ssh/id_ed25519", // Standard Ed25519 key
		os.Getenv("HOME") + "/.ssh/id_ecdsa",   // Standard ECDSA key
		os.Getenv("HOME") + "/.ssh/id_dsa",     // Standard DSA key
	}

	var signers []ssh.Signer

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			// Key file exists, try to load it
			key, err := os.ReadFile(path)
			if err != nil {
				continue // Try next key
			}

			signer, err := ssh.ParsePrivateKey(key)
			if err != nil {
				continue // Try next key
			}

			signers = append(signers, signer)
		}
	}

	return signers, nil
}

// getPassword prompts for SSH password if key authentication fails
func (d *Deployer) getPassword() (string, error) {
	// In a real implementation, you might want to prompt for password,
	// but for now we'll return an empty string which will cause fallback
	// to key-based authentication.
	return "", nil
}