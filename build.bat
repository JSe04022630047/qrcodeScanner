:: this is just for fancy "about" commit message
go build -ldflags "-X main.GitCommit=$(git rev-parse --short HEAD)" -o QRScan.exe