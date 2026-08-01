# finalize-mvp.ps1
# Finishes the t3.small node swap, gets every pod healthy, verifies HTTPS,
# and installs HPA for api/web. Run from anywhere (PowerShell 5.1+).
$ErrorActionPreference = 'Stop'

# Make sure AWS CLI is on PATH for kubectl's credential plugin
$env:Path += ';C:\Program Files\Amazon\AWSCLIV2'

$here  = $PSScriptRoot                      # ...\infrastructure\scripts
$repo  = Split-Path -Parent (Split-Path -Parent $here)
$infra = Join-Path $repo 'infrastructure'
$kube  = Join-Path $repo 'kubernetes'

$APP_URL = 'https://35.175.60.72.sslip.io'

function Wait-AllPodsRunning {
    param([int]$Tries = 40)
    $i = 0
    do {
        Start-Sleep -Seconds 15
        $i++
        $bad = @(kubectl get pods -A --no-headers 2>&1 | Select-String -NotMatch 'Running|Completed')
        Write-Host "  [$i/$Tries] non-ready pods: $($bad.Count)"
        if ($bad.Count -eq 0) { return $true }
    } while ($i -lt $Tries)
    return ($bad.Count -eq 0)
}

Write-Host "===== 1/6 Terraform: swap node group to t3.small (10-20 min) ====="
Push-Location $infra
try {
    cmd /c "terraform apply -auto-approve -replace=module.eks.aws_eks_node_group.main -lock-timeout=10m -no-color"
    if ($LASTEXITCODE -ne 0) { throw "terraform apply failed (exit $LASTEXITCODE)" }
} finally { Pop-Location }

Write-Host "===== 2/6 Wait for all 3 nodes Ready, maxPods=9 ====="
$i = 0
do {
    Start-Sleep -Seconds 20
    $i++
    $nodes = kubectl get nodes --no-headers 2>&1 | Out-String
    Write-Host "----- nodes (try $i):"
    Write-Host $nodes
    $ready = @(kubectl get nodes --no-headers 2>&1 | Select-String ' Ready')
    $podsCap = @(kubectl get nodes --no-headers 2>&1 | Select-String 'Ready' | Select-String -NotMatch 'NotReady')
} while (($ready.Count -lt 3) -and $i -lt 30)

Write-Host "===== 3/6 Reschedule stuck pods (CNI had old-node IPs) ====="
$stuck = @(kubectl get pods -A --no-headers 2>&1 | Select-String -NotMatch 'Running|Completed')
foreach ($line in $stuck) {
    $f = ($line -split '\s+')
    if ($f.Count -ge 2) {
        Write-Host "  deleting $($f[0])/$($f[1])"
        kubectl delete pod -n $f[0] $f[1] --wait=false 2>&1 | Out-Null
    }
}
if (-not (Wait-AllPodsRunning)) {
    Write-Host "WARN: some pods still not Running after timeout. Check: kubectl get pods -A"
}

Write-Host "===== 4/6 Verify HTTPS + API ====="
$code = curl.exe -sS -m 20 "$APP_URL/" -o NUL -w '%{http_code}'
Write-Host "HTTPS / status: $code"
$api = curl.exe -sS -m 20 "$APP_URL/api/v1/projects"
Write-Host "HTTPS /api/v1/projects: $api"

Write-Host "===== 5/6 Metrics server + HPA ====="
kubectl top nodes 2>&1
kubectl apply -f (Join-Path $kube 'hpa\api-hpa.yaml')
kubectl apply -f (Join-Path $kube 'hpa\web-hpa.yaml')
kubectl get hpa -n devplatform

Write-Host "===== 6/6 DONE ====="
Write-Host "App URL : $APP_URL"
Write-Host "Nodes   :" (kubectl get nodes -o wide --no-headers | Out-String)
