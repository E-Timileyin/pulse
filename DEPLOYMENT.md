# Deployment Guide

Pulse can be deployed as a headless backend server on a VPS, making it accessible via a web interface or API.

## 1. Build for Linux (VPS)

If you are developing on a different OS, cross-compile for Linux:

```bash
GOOS=linux GOARCH=amd64 go build -o pulse-linux-amd64 .
```

## 2. Deploy to VPS

Copy the binary to your server:

```bash
scp pulse-linux-amd64 user@your-vps-ip:/usr/local/bin/pulse
```

## 3. Serving a React Frontend

You can serve your React application directly from Pulse using the `--static` flag.

1.  Build your React app: `npm run build`
2.  Copy the `dist` or `build` folder to your VPS (e.g., `/var/www/pulse-ui`).

## 4. Running Pulse Server

Run Pulse in server mode:

```bash
pulse server --host 0.0.0.0 --port 8080 --static /var/www/pulse-ui
```

Your React app will be available at `http://your-vps-ip:8080/`, and the API at `http://your-vps-ip:8080/download`.

## 5. Systemd Service (Recommended)

Create a systemd service to keep Pulse running in the background.

File: `/etc/systemd/system/pulse.service`

```ini
[Unit]
Description=Pulse Download Manager
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/pulse server --host 0.0.0.0 --port 8080 --static /var/www/pulse-ui
Restart=on-failure
Environment=HOME=/root

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl enable pulse
sudo systemctl start pulse
```

## 6. Accessing the API from React

In your React app, you can send download requests to the backend:

```javascript
fetch("/download", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    url: "https://youtube.com/watch?v=...",
    quality: "1080p", // Optional
  }),
});
```

The server sends CORS headers by default to allow cross-origin requests if you decide to host the frontend separately.
