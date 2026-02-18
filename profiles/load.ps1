$count = 5000;

Write-Host "Running load test... ($count requests)" -ForegroundColor Cyan

if (-not (Get-Command "curl" -ErrorAction SilentlyContinue)) {
    Write-Warning "curl NOT FOUND."
    exit 1
}

$parallelLimit = 50
$jobs = @()
$batchSize = $parallelLimit

for ($i = 1; $i -le $count; $i++) {
    while ((Get-Job -State Running).Count -ge $parallelLimit) {
        Start-Sleep -Milliseconds 10
    }

    $url = "https://example$i.com/path?q=$i"
    
    $body = "{`\`"url`\`": `\`"$url`\`"}"

    $job = Start-Job -ScriptBlock {
        param($uri, $body)
        curl.exe -s -X POST "$uri" -H "Content-Type: application/json" -d "$body" *> $null
    } -ArgumentList @("http://localhost:8080/api/shorten", $body)

    $jobs += $job

    if ($i % 100 -eq 0) {
        Write-Host "Sent $i requests..."
    }
}

Wait-Job -Job $jobs | Out-Null

Remove-Job -Job $jobs

Write-Host "Load sent. Done $count requests" -ForegroundColor Green
