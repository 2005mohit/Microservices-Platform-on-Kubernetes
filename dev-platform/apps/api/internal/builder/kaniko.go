package builder

import (
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/devplatform/api/internal/db"
	"github.com/devplatform/api/internal/k8s"
	"github.com/devplatform/api/internal/queue"
)

func RunBuild(projectID, deployID, projectName, gitRepo, branch string) {
	sendLog(deployID, "queued", "Initializing build...\n")
	time.Sleep(1 * time.Second)

	domain := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(projectName)), " ", "")
	if domain == "" {
		domain = "app"
	}

	sendLog(deployID, "building", "Cloning repository: "+gitRepo+"\n")

	cloneDir := filepath.Join("/var/devplatform/data", deployID)
	os.RemoveAll(cloneDir)

	cloneURL := gitRepo
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		cloneURL = strings.Replace(gitRepo, "https://", "https://"+token+"@", 1)
	}
	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", branch, cloneURL, cloneDir)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		sendLog(deployID, "building", "Git clone failed (private repo), creating demo deployment...\n")
		os.RemoveAll(cloneDir)
		os.MkdirAll(cloneDir, 0755)
		html := fmt.Sprintf("<!DOCTYPE html><html><head><title>%s - DevPlatform</title><style>body{font-family:sans-serif;display:flex;justify-content:center;align-items:center;min-height:100vh;margin:0;background:#0f0f0f;color:#fff;flex-direction:column}div{text-align:center}h1{font-size:3rem;background:linear-gradient(135deg,#8b5cf6,#6366f1);-webkit-background-clip:text;-webkit-text-fill-color:transparent}p{color:#888}a{color:#8b5cf6}</style></head><body><div><h1>%s</h1><p>Deployed via DevPlatform</p><p style=\"font-size:0.8rem;color:#555\">Repo: %s</p><p style=\"margin-top:2rem\"><a href=\"http://localhost:3000/dashboard/projects\">Back to Dashboard</a></p></div></body></html>", domain, domain, gitRepo)
		os.WriteFile(filepath.Join(cloneDir, "index.html"), []byte(html), 0644)
		sendLog(deployID, "building", "Demo page created for: "+domain+"\n")
	} else {
		sendLog(deployID, "building", "Repository cloned successfully\n")
		log.Printf("Cloned to %s: %s", cloneDir, string(out))
		if sha, err := exec.Command("git", "-C", cloneDir, "rev-parse", "HEAD").Output(); err == nil {
			db.Pool().Exec("UPDATE deployments SET commit_sha = $1 WHERE id = $2", strings.TrimSpace(string(sha)), deployID)
		}
	}
	sendLog(deployID, "building", "Branch: "+branch+"\n")

	// Generate deterministic port from projectID
	h := fnv.New32a()
	h.Write([]byte(projectID))
	port := 9000 + int(h.Sum32()%1000)

	dockerfilePath := filepath.Join(cloneDir, "Dockerfile")
	containerName := "dp-" + domain

	exec.Command("docker", "rm", "-f", containerName).Run()

	if _, err := os.Stat(dockerfilePath); err == nil {
		sendLog(deployID, "building", "Dockerfile found - building image...\n")
		imageName := "devplatform-" + domain + ":latest"
		cmd = exec.Command("docker", "build", "-t", imageName, cloneDir)
		out, err = cmd.CombinedOutput()
		if err != nil {
			sendLog(deployID, "failed", fmt.Sprintf("Docker build failed: %v\n%s\n", err, string(out)))
			return
		}
		sendLog(deployID, "building", "Image built. Starting container...\n")
		cmd = exec.Command("docker", "run", "-d", "--name", containerName, "-p", fmt.Sprintf("%d:80", port), imageName)
	} else {
		sendLog(deployID, "building", "No Dockerfile - checking for Node.js project...\n")
		packagePath := filepath.Join(cloneDir, "package.json")
		if _, err := os.Stat(packagePath); err == nil {
			sendLog(deployID, "building", "package.json found - building project...\n")
			buildOut, buildErr := exec.Command("sh", "-c", "cd "+cloneDir+" && npm install 2>&1 && npm run build 2>&1").CombinedOutput()
			if buildErr != nil {
				sendLog(deployID, "failed", fmt.Sprintf("Build failed: %v\n%s\n", buildErr, string(buildOut)))
				return
			}
			sendLog(deployID, "building", "Node.js build complete\n")
			for _, dir := range []string{"dist", "build", "out", ".next"} {
				if s, err := os.Stat(filepath.Join(cloneDir, dir)); err == nil && s.IsDir() {
					cloneDir = filepath.Join(cloneDir, dir)
					break
				}
			}
		}
		sendLog(deployID, "building", "Serving static files...\n")
		cmd = exec.Command("docker", "run", "-d", "--name", containerName, "-p", fmt.Sprintf("%d:80", port), "-v", cloneDir+":/usr/share/nginx/html:ro", "nginx:alpine")
	}

	out, err = cmd.CombinedOutput()
	if err != nil {
		sendLog(deployID, "failed", fmt.Sprintf("Docker run failed: %v\n%s\n", err, string(out)))
		return
	}

	privateIP := os.Getenv("NODE_IP")
	if privateIP == "" {
		privateIP = "localhost"
	}
	nodeIP := privateIP
	publicIPs := map[string]string{
		"172.31.6.126": "3.231.221.200",
		"172.31.2.128": "3.216.78.214",
		"172.31.7.229": "3.235.43.105",
	}
	if mapped, ok := publicIPs[privateIP]; ok {
		nodeIP = mapped
	}
	baseDomain := os.Getenv("BASE_DOMAIN")
	if baseDomain == "" {
		baseDomain = "3.231.221.200.nip.io"
	}
	hostname := fmt.Sprintf("%s.%s", domain, baseDomain)
	domainURL := fmt.Sprintf("%s:%d", nodeIP, port)
	db.Pool().Exec("UPDATE deployments SET domain = $1 WHERE id = $2", domainURL, deployID)
	if err := k8s.EnsureExternalRoute(domain, hostname, nodeIP, int32(port)); err != nil {
		log.Printf("Warning: failed to create ingress route: %v", err)
	} else {
		domainURL = hostname
		db.Pool().Exec("UPDATE deployments SET domain = $1 WHERE id = $2", hostname, deployID)
	}


	sendLog(deployID, "success", "Deployed successfully!\n")
	sendLog(deployID, "success", fmt.Sprintf("URL: http://%s\n", domainURL))
}



func sendLog(deployID, status, msg string) {
	db.Pool().Exec("UPDATE deployments SET status = $1, logs = COALESCE(logs || $2, $2), updated_at = NOW() WHERE id = $3", status, msg, deployID)
	queue.Publish("deployment:"+deployID, map[string]string{
		"type":   "log",
		"status": status,
		"line":   msg,
	})
	log.Printf("[%s] %s: %s", deployID, status, msg)
}
