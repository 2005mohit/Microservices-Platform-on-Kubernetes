package builder

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/devplatform/api/internal/db"
	"github.com/devplatform/api/internal/k8s"
	"github.com/devplatform/api/internal/queue"
)

const (
	buildNamespace = "devplatform"
	buildTimeout   = 12 * time.Minute
)

var (
	sanitizeRe = regexp.MustCompile(`[^a-z0-9-]`)
	exposeRe   = regexp.MustCompile(`(?i)^\s*EXPOSE\s+(\d+)`)
)

// prepareScript runs in the build Job's init container. It clones the repo and
// generates a Dockerfile when the project has no Dockerfile of its own.
const prepareScript = `set -e
rm -rf /workspace/*
URL="$REPO_URL"
if [ -n "$GITHUB_TOKEN" ]; then
  URL=$(echo "$REPO_URL" | sed "s|https://|https://${GITHUB_TOKEN}@|")
fi
git clone --depth 1 --branch "$BRANCH" "$URL" /workspace
cd /workspace
if [ "$STRATEGY" = "node" ]; then
cat > /workspace/Dockerfile <<'DOCKERFILE'
FROM node:20-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci || npm install
COPY . .
RUN npm run build || true
FROM node:20-alpine
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app/package*.json ./
RUN npm ci --omit=dev || npm install --omit=dev
COPY --from=build /app ./
EXPOSE 3000
CMD ["sh", "-c", "npm start || npm run start || npx --yes serve -s . -l 3000"]
DOCKERFILE
elif [ "$STRATEGY" = "static" ]; then
cat > /workspace/Dockerfile <<'DOCKERFILE'
FROM nginx:alpine
COPY . /usr/share/nginx/html
DOCKERFILE
fi
`

func RunBuild(projectID, deployID, projectName, gitRepo, branch string) {
	sendLog(deployID, "queued", "Initializing build...\n")

	cloneDir, err := cloneAndInspect(deployID, gitRepo, branch)
	if err != nil {
		sendLog(deployID, "failed", fmt.Sprintf("Failed to clone repository: %v\n", err))
		return
	}

	strategy, port := inspectProject(cloneDir)
	os.RemoveAll(cloneDir)

	registry := getEnv("REGISTRY", "ghcr.io/2005mohit")
	image := fmt.Sprintf("%s/dp-%s:%s", registry, sanitize(projectName), deployID[:8])

	// Make sure the GHCR push/pull secret exists in the build namespace.
	if err := k8s.EnsurePullSecret(buildNamespace); err != nil {
		sendLog(deployID, "failed", fmt.Sprintf("Failed to prepare registry auth: %v\n", err))
		return
	}

	jobName := "build-" + deployID[:8]
	sendLog(deployID, "building", fmt.Sprintf("Cloning done. Building image %s with Kaniko...\n", image))
	if err := createBuildJob(jobName, gitRepo, branch, strategy, image); err != nil {
		sendLog(deployID, "failed", fmt.Sprintf("Failed to create build job: %v\n", err))
		return
	}

	if !watchBuildJob(deployID, jobName) {
		sendLog(deployID, "failed", "Build failed - see logs above\n")
		return
	}

	name := sanitize(projectName)
	hostname := fmt.Sprintf("%s.%s", name, getEnv("BASE_DOMAIN", "localhost"))

	sendLog(deployID, "building", fmt.Sprintf("Image built. Deploying to namespace tenant-%s...\n", name))
	if err := k8s.DeployUserApp(name, hostname, image, port); err != nil {
		sendLog(deployID, "failed", fmt.Sprintf("Deploy failed: %v\n", err))
		return
	}

	url := fmt.Sprintf("https://%s", hostname)
	if _, err := db.Pool().Exec(`UPDATE deployments SET domain=$1, updated_at=NOW() WHERE id=$2`, url, deployID); err != nil {
		log.Printf("failed to update deployment domain: %v", err)
	}
	sendLog(deployID, "success", "Deployed successfully!\nURL: "+url+"\n")
}

func cloneAndInspect(deployID, gitRepo, branch string) (string, error) {
	cloneDir, err := os.MkdirTemp("", "dpbuild-"+deployID[:8])
	if err != nil {
		return "", err
	}
	cloneURL := gitRepo
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		cloneURL = strings.Replace(gitRepo, "https://", "https://"+token+"@", 1)
	}
	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", branch, cloneURL, cloneDir)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(cloneDir)
		return "", fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return cloneDir, nil
}

// inspectProject decides the build strategy and the container port.
func inspectProject(dir string) (strategy string, port int32) {
	if _, err := os.Stat(filepath.Join(dir, "Dockerfile")); err == nil {
		if f, err := os.Open(filepath.Join(dir, "Dockerfile")); err == nil {
			sc := bufio.NewScanner(f)
			for sc.Scan() {
				if m := exposeRe.FindStringSubmatch(sc.Text()); m != nil {
					if p, err := strconv.Atoi(m[1]); err == nil && p > 0 && p < 65536 {
						f.Close()
						return "dockerfile", int32(p)
					}
				}
			}
			f.Close()
		}
		return "dockerfile", 80
	}
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
		return "node", 3000
	}
	return "static", 80
}

func createBuildJob(jobName, gitRepo, branch, strategy, image string) error {
	env := []corev1.EnvVar{
		{Name: "REPO_URL", Value: gitRepo},
		{Name: "BRANCH", Value: branch},
		{Name: "STRATEGY", Value: strategy},
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		env = append(env, corev1.EnvVar{
			Name: "GITHUB_TOKEN",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "github-pat"},
					Key:                  "GITHUB_TOKEN",
				},
			},
		})
	}

	backoff := int32(0)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: jobName, Namespace: buildNamespace},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoff,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					InitContainers: []corev1.Container{
						{
							Name:    "prepare",
							Image:   "alpine/git:2.45.2",
							Command: []string{"/bin/sh", "-c", prepareScript},
							Env:     env,
							VolumeMounts: []corev1.VolumeMount{
								{Name: "workspace", MountPath: "/workspace"},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "kaniko",
							Image: "gcr.io/kaniko-project/executor:v1.23.2",
							Args: []string{
								"--context=dir:///workspace",
								"--dockerfile=/workspace/Dockerfile",
								"--destination=" + image,
								"--cache=true",
								"--cache-repo=" + getEnv("REGISTRY", "ghcr.io/2005mohit") + "/kaniko-cache",
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "workspace", MountPath: "/workspace"},
								{Name: "docker-config", MountPath: "/kaniko/.docker"},
							},
						},
					},
					Volumes: []corev1.Volume{
						{Name: "workspace", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
						{
							Name: "docker-config",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "ghcr-push",
									Items:      []corev1.KeyToPath{{Key: ".dockerconfigjson", Path: "config.json"}},
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := k8s.Client.BatchV1().Jobs(buildNamespace).Create(context.Background(), job, metav1.CreateOptions{})
	return err
}

// watchBuildJob polls the Job until it completes, fails, or times out. It
// streams the build logs into the deployment's log buffer on completion.
func watchBuildJob(deployID, jobName string) bool {
	ctx := context.Background()
	deadline := time.Now().Add(buildTimeout)

	for time.Now().Before(deadline) {
		job, err := k8s.Client.BatchV1().Jobs(buildNamespace).Get(ctx, jobName, metav1.GetOptions{})
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		for _, cond := range job.Status.Conditions {
			if cond.Type == batchv1.JobComplete && cond.Status == corev1.ConditionTrue {
				streamJobLogs(deployID, jobName, "prepare")
				streamJobLogs(deployID, jobName, "kaniko")
				return true
			}
			if cond.Type == batchv1.JobFailed && cond.Status == corev1.ConditionTrue {
				streamJobLogs(deployID, jobName, "prepare")
				streamJobLogs(deployID, jobName, "kaniko")
				return false
			}
		}
		time.Sleep(3 * time.Second)
	}
	sendLog(deployID, "building", "Build timed out\n")
	return false
}

func streamJobLogs(deployID, jobName, container string) {
	ctx := context.Background()
	pods, err := k8s.Client.CoreV1().Pods(buildNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: "job-name=" + jobName,
	})
	if err != nil {
		return
	}
	for _, p := range pods.Items {
		req := k8s.Client.CoreV1().Pods(buildNamespace).GetLogs(p.Name, &corev1.PodLogOptions{Container: container})
		if body, err := req.DoRaw(ctx); err == nil {
			sendLog(deployID, "building", string(body))
		}
	}
}

func sendLog(deployID, status, msg string) {
	if _, err := db.Pool().Exec(
		`UPDATE deployments SET status = $1, logs = COALESCE(logs || $2, $2), updated_at = NOW() WHERE id = $3`,
		status, msg, deployID,
	); err != nil {
		log.Printf("failed to update deployment logs: %v", err)
	}
	queue.Publish("deployment:"+deployID, map[string]string{
		"type":   "log",
		"status": status,
		"line":   msg,
	})
	log.Printf("[%s] %s: %s", deployID, status, msg)
}

func sanitize(name string) string {
	s := sanitizeRe.ReplaceAllString(strings.ToLower(strings.TrimSpace(name)), "-")
	if len(s) > 40 {
		s = s[:40]
	}
	return strings.Trim(s, "-")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
