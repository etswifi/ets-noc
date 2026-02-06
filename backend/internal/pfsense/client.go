package pfsense

import (
	"context"
	"encoding/xml"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type DHCPStaticMapping struct {
	Hostname string
	IPAddr   string
	MAC      string
}

type Client struct {
	host     string
	port     int
	username string
	password string
}

func NewClient(host string, port int, username, password string) *Client {
	return &Client{
		host:     host,
		port:     port,
		username: username,
		password: password,
	}
}

// GetDHCPStaticMappings fetches DHCP static mappings from pfSense
func (c *Client) GetDHCPStaticMappings(ctx context.Context) ([]DHCPStaticMapping, error) {
	config := &ssh.ClientConfig{
		User: c.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Execute command to get config.xml content with DHCP static mappings
	// pfSense stores DHCP config in /cf/conf/config.xml
	cmd := `cat /cf/conf/config.xml | grep -A 5 "<staticmap>" | grep -E "(hostname|ipaddr|mac)" | sed 's/<[^>]*>//g' | sed 's/^[ \t]*//'`

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return parseStaticMappings(string(output)), nil
}

// parseStaticMappings parses the grep output into DHCPStaticMapping structs
func parseStaticMappings(output string) []DHCPStaticMapping {
	var mappings []DHCPStaticMapping
	lines := strings.Split(strings.TrimSpace(output), "\n")

	var current DHCPStaticMapping
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Pattern: hostname, ipaddr, mac repeating
		switch i % 3 {
		case 0:
			current.Hostname = line
		case 1:
			current.IPAddr = line
		case 2:
			current.MAC = line
			// Complete mapping, add to list
			mappings = append(mappings, current)
			current = DHCPStaticMapping{}
		}
	}

	return mappings
}

// Alternative method using XML parsing (more robust)
type ConfigXML struct {
	DHCPd struct {
		LAN struct {
			StaticMaps []struct {
				MAC      string `xml:"mac"`
				IPAddr   string `xml:"ipaddr"`
				Hostname string `xml:"hostname"`
			} `xml:"staticmap"`
		} `xml:"lan"`
	} `xml:"dhcpd"`
}

// GetDHCPStaticMappingsXML fetches DHCP static mappings using XML parsing
func (c *Client) GetDHCPStaticMappingsXML(ctx context.Context) ([]DHCPStaticMapping, error) {
	config := &ssh.ClientConfig{
		User: c.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Get the full config.xml
	output, err := session.CombinedOutput("cat /cf/conf/config.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	var cfg ConfigXML
	if err := xml.Unmarshal(output, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	var mappings []DHCPStaticMapping
	for _, sm := range cfg.DHCPd.LAN.StaticMaps {
		mappings = append(mappings, DHCPStaticMapping{
			Hostname: sm.Hostname,
			IPAddr:   sm.IPAddr,
			MAC:      sm.MAC,
		})
	}

	return mappings, nil
}

// DetermineDeviceType returns the device type based on IP address
func DetermineDeviceType(ipAddr string) string {
	parts := strings.Split(ipAddr, ".")
	if len(parts) != 4 {
		return "Unknown"
	}

	lastOctet, err := strconv.Atoi(parts[3])
	if err != nil {
		return "Unknown"
	}

	if lastOctet == 1 {
		return "Router"
	} else if lastOctet >= 100 {
		return "Switch"
	} else if lastOctet >= 2 && lastOctet <= 99 {
		return "WAP"
	}

	return "Unknown"
}
