# FullStackApp01
Full stack app using Antigravity IDE &amp; AI


## Installation notes

On the target VM the following steps should be completed:

### 1. Prerequisites on the VM
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Git
sudo apt install -y git

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

### 4. Setup Systemd Service (Backend and Frontend)
```bash
cd ~/app/scripts
./create_backend_service.sh
./create_frontend_service.sh
```

### 5. Cloudflare setup

#### 5.1. Installing cloudflared

```bash
cd
sudo mkdir -p --mode=0755 /usr/share/keyrings
curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | sudo tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null
echo 'deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared jammy main' | sudo tee /etc/apt/sources.list.d/cloudflared.list
sudo apt-get update && sudo apt-get install cloudflared
```

#### 5.2. Authenticate
```bash
cloudflared tunnel login
```
Follow the URL to authorize your domain.

#### 5.3. Create Tunnel
```bash
cloudflared tunnel create fullstack-vm
# Note the UUID and credentials path outputted.
```

#### 5.4. Configure Tunnel

```bash
# Create/edit the config file 
nano ~/.cloudflared/config.yml
```
Suppose the domain is set to xxx.yyy.zzz

```yaml
tunnel: <YOUR_TUNNEL_UUID>
credentials-file: /home/ubuntu/.cloudflared/<YOUR_TUNNEL_UUID>.json

ingress:
  # API Routes -> Go Backend (port 8080)
  - hostname: xxx.yyy.zzz
    path: /login
    service: http://localhost:8080
  - hostname: xxx.yyy.zzz
    path: /register
    service: http://localhost:8080
  - hostname: xxx.yyy.zzz
    path: /counter
    service: http://localhost:8080
  - hostname: xxx.yyy.zzz
    path: /change-password
    service: http://localhost:8080

  # All other routes -> React Frontend (port 5173)
  - hostname: xxx.yyy.zzz
    service: http://localhost:5173

  # Catch-all
  - service: http_status:404
```

#### 5.5. DNS Routing
```bash
cloudflared tunnel route dns fullstack-vm xxx.yyy.zzz
```

#### 5.6. Start Tunnel Service
```bash
# Create the system directory
sudo mkdir -p /etc/cloudflared

# Copy your config and the credentials JSON file
sudo cp ~/.cloudflared/config.yml /etc/cloudflared/
sudo cp ~/.cloudflared/*.json /etc/cloudflared/

sudo nano /etc/cloudflared/config.yml
# Adjust the credentials-file to /etc/cloudflared/<YOUR_TUNNEL_UUID>.json

sudo cloudflared service install
sudo systemctl start cloudflared
```

#### 5.7. Additional settings
Add allowed sites in `~/app/frontend/vite.config.ts`. Something like:
```typescript
// https://vite.dev/config/
export default defineConfig({
    plugins: [react()],
    server: {
        allowedHosts: ['xxx.yyy.zzz']
    }
})
```

then git-pull the solution on the VM and execute:
```bash
sudo systemctl restart app-frontend
```

#### 5.8. Troubleshooting
- **Frontend Errors**: Check if API calls in `App.tsx` / `Login.tsx` are relative (e.g. `/login`, not `http://localhost:8080/login`).
- **502 Bad Gateway**: Check if backend/frontend services are running (`systemctl status app frontend`).