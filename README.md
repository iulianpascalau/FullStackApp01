# FullStackApp01
Full stack app using Antigravity IDE &amp; AI


## Installation notes

On the target VM the following steps should be completed:

### 1. Prerequisites on the VM
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Nginx and Git
sudo apt install -y nginx git

# Install Go (Adjust version if needed, your go.mod says 1.24 so you need a very recent version)
GO_LATEST_TESTED="1.24.11"
ARCH=$(dpkg --print-architecture)
wget https://dl.google.com/go/go${GO_LATEST_TESTED}.linux-${ARCH}.tar.gz
sudo tar -C /usr/local -xzf go${GO_LATEST_TESTED}.linux-${ARCH}.tar.gz
rm go${GO_LATEST_TESTED}.linux-${ARCH}.tar.gz

echo "export GOPATH=$HOME/go" >> ~/.profile
echo "export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin" >> ~/.profile
echo "export GOPATH=$HOME/go" >> ~/.profile
source ~/.profile
go version

# Install Node.js (for building frontend)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
```

### 2. Clone & Build the Application
```bash
# Clone the repository (replace with your actual repo URL)
cd ~
git clone https://github.com/iulianpascalau/FullStackApp01.git app
cd app

# Ensure you are on the main branch (after you merge your PR)
git checkout main
git pull origin main

# --- Build Backend ---
# Create a binary named 'server'
go build -o server main.go

# --- Build Frontend ---
cd frontend
npm install
npm run build
# This creates a 'dist' folder with your static site
```

### 3. Configure Environment Variables
```bash
cd ~/app
# Copy and use the example .env file 
cp .env.example .env
nano .env
```

### 4. Setup Systemd Service (Backend)
```bash
cd ~/app/scripts
./create_service.sh
```

