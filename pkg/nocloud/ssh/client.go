package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	dialTimeout     = 30 * time.Second
	keepAlivePeriod = 30 * time.Second
)

type Client struct {
	mu     sync.Mutex
	client *gossh.Client
	dial   func() (*gossh.Client, error)
}

func NewSSHClient(sshHost, sshUser, keyPath string) (*Client, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("private key error: %w", err)
	}

	signer, err := gossh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("private key parse error: %w", err)
	}

	auth := []gossh.AuthMethod{gossh.PublicKeys(signer)}
	dial := func() (*gossh.Client, error) {
		return dialClient(sshHost, sshUser, auth)
	}

	client, err := dial()
	if err != nil {
		return nil, err
	}

	return &Client{client: client, dial: dial}, nil
}

func NewSSHClientFromPassword(sshHost, sshUser, pass string) (*Client, error) {
	auth := []gossh.AuthMethod{gossh.Password(pass)}
	dial := func() (*gossh.Client, error) {
		return dialClient(sshHost, sshUser, auth)
	}

	client, err := dial()
	if err != nil {
		return nil, err
	}

	return &Client{client: client, dial: dial}, nil
}

func dialClient(host, user string, auth []gossh.AuthMethod) (*gossh.Client, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user home directory: %w", err)
	}

	hostKeyCallback, err := knownhosts.New(homeDir + "/.ssh/known_hosts")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize known_hosts callback: %w", err)
	}

	config := &gossh.ClientConfig{
		User:            user,
		Auth:            auth,
		HostKeyCallback: hostKeyCallback,
		Timeout:         dialTimeout,
	}

	conn, err := net.DialTimeout("tcp", host, dialTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	if tc, ok := conn.(*net.TCPConn); ok {
		_ = tc.SetKeepAlive(true)
		_ = tc.SetKeepAlivePeriod(keepAlivePeriod)
	}

	cc, chans, reqs, err := gossh.NewClientConn(conn, host, config)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to establish ssh session: %w", err)
	}

	return gossh.NewClient(cc, chans, reqs), nil
}

func (s *Client) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == nil {
		return nil
	}
	err := s.client.Close()
	s.client = nil
	return err
}

func (s *Client) reconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		_ = s.client.Close()
		s.client = nil
	}

	client, err := s.dial()
	if err != nil {
		return err
	}
	s.client = client
	return nil
}

func (s *Client) RunCommand(command string) (string, error) {
	out, err := s.runCommandOnce(command)
	if err == nil || !isReconnectable(err) {
		return out, err
	}

	if reconnectErr := s.reconnect(); reconnectErr != nil {
		return "", fmt.Errorf("ssh reconnect failed: %w (original: %w)", reconnectErr, err)
	}

	return s.runCommandOnce(command)
}

func (s *Client) runCommandOnce(command string) (string, error) {
	s.mu.Lock()
	client := s.client
	s.mu.Unlock()

	if client == nil {
		return "", errors.New("ssh client is not connected")
	}

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create new session: %w", err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	if err := session.Run(command); err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	return b.String(), nil
}

func isReconnectable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	msg := strings.ToLower(err.Error())
	for _, sub := range []string{
		"timeout",
		"timed out",
		"connection reset",
		"broken pipe",
		"eof",
		"client is closed",
		"use of closed network connection",
		"connection refused",
		"i/o timeout",
		"no route to host",
		"network is unreachable",
	} {
		if strings.Contains(msg, sub) {
			return true
		}
	}
	return false
}
