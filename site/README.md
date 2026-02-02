# Web Assembly Project for Dizz CLI

This directory contains a complete WebAssembly-based web interface for your Dizz CLI tool, independent of the main project.

## 🌟 Features

- **Version Display**: Automatically fetches latest version from git tags
- **CLI Commands**: Interactive command documentation
- **One-Click Installation**: Platform-specific install scripts with curl support
- **Responsive Design**: Modern, mobile-friendly interface
- **WASM Integration**: WebAssembly components for advanced functionality
- **API Endpoints**: RESTful API for version and command information

## 📁 Structure

```
site/
├── README.md                # Project documentation
├── server/                  # Go HTTP server
│   └── main.go             # Main server with API endpoints
├── public/                  # Static assets
│   ├── index.html           # Landing page
│   ├── styles.css           # Modern styling
│   ├── favicon.ico          # Site icon
│   └── assets/              # Additional assets
├── wasm/                    # WebAssembly components
│   └── main.go             # WASM entrypoint
├── js/                      # Frontend JavaScript
│   ├── app.js              # Main application logic
│   └── install.js          # Installation script handling
├── scripts/                 # Install scripts
│   ├── install.sh           # Unix/macOS installer
│   └── install.ps1          # Windows PowerShell installer
└── docs/                    # CLI documentation
    ├── index.md             # Documentation overview
    ├── install.md           # Installation guide
    ├── commands.md          # Command reference
    ├── workflow.md          # Development workflow
    └── philosophy.md        # Project philosophy
```

## 🚀 Quick Start

### 1. Build the WebAssembly Project

```bash
# Run the build script
./scripts/build-site.sh
```

### 2. Start the Web Server

```bash
cd site/server
go run main.go
```

The server will start on port 8080 (configurable via PORT environment variable).

### 3. Access the Web Interface

Visit http://localhost:8080 to see the landing page.

## 🔧 Configuration

### Environment Variables

- `PORT`: Server port (default: 8080)
- `DIZZ_REPO`: GitHub repository for releases (default: yourusername/dizz)

### Customization

- Update `site/public/index.html` for branding and content
- Modify `site/public/styles.css` for styling
- Extend `site/server/main.go` for additional API endpoints
- Update install scripts with your actual repository information

## 📡 API Endpoints

- `GET /api/version` - Latest version information from git tags
- `GET /api/commands` - CLI command documentation
- `GET /install.sh` - Unix/macOS installation script
- `GET /install.ps1` - Windows installation script

## 🔒 Security Notes

The install scripts download binaries from GitHub releases. Ensure:

1. Your repository uses proper release signatures
2. Download URLs are validated
3. Binary verification is implemented in production

## 🌐 Deployment

### Static Site Deployment

For static hosting (GitHub Pages, Netlify, etc.):

1. Build the WASM module using `./scripts/build-site.sh`
2. Deploy the `site/public/` directory
3. Update install script URLs in the HTML

### Server Deployment

For server deployment:

1. Build the server binary
2. Run with appropriate configuration
3. Set up reverse proxy if needed

## 🔄 Version Management

The web interface automatically detects and displays the latest version:

- Fetches from git tags via `git describe --tags --abbrev=0`
- Integrates with GitHub releases for downloads
- Supports GoReleaser workflow for automated releases

## 🎨 Design Features

- **Modern UI**: Clean, professional design with smooth animations
- **Mobile Responsive**: Works perfectly on all devices
- **Dark Mode Ready**: CSS variables support theming
- **Accessibility**: Semantic HTML and keyboard navigation
- **Progressive Enhancement**: Works with/without JavaScript/WASM

## 🛠 Development

### Local Development

```bash
# Install dependencies
go mod tidy

# Build and run
cd site/server && go run main.go

# Watch for changes (optional)
air -c .air.toml  # if using air for hot reload
```

### WebAssembly Development

```bash
# Build WASM only
cd site/wasm
GOOS=js GOARCH=wasm go build -o ../public/dizz.wasm .

# Test WASM in browser
python3 -m http.server 8080  # Simple HTTP server
```

## 📚 Integration with GoReleaser

The install scripts are designed to work with GoReleaser:

```yaml
# .goreleaser.yaml
release:
  github:
    owner: yourusername
    name: dizz

builds:
  - env: [CGO_ENABLED=0]
    goos: [linux, windows, darwin]
    goarch: [amd64, arm64]
```

## 🤝 Contributing

1. Update the web content as needed
2. Test install scripts across platforms
3. Verify API endpoints work correctly
4. Check mobile responsiveness

## 📄 License

This web interface follows the same license as the main Dizz project.

---

**Note**: Remember to update the repository URLs and branding information to match your actual project before deploying to production.
