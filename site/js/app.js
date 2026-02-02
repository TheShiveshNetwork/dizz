// Main application JavaScript
class DizzApp {
    constructor() {
        this.version = null;
        this.commands = [];
        this.wasmReady = false;
        this.init();
    }

    async init() {
        await this.loadVersion();
        await this.loadCommands();
        this.setupEventListeners();
        
        // Try to initialize WASM
        try {
            await this.initWasm();
        } catch (error) {
            console.log('WASM not available, falling back to API');
        }
    }

    async loadVersion() {
        try {
            const response = await fetch('/api/version');
            const data = await response.json();
            this.version = data;
            this.updateVersionDisplay();
        } catch (error) {
            console.error('Failed to load version:', error);
            document.getElementById('version').textContent = 'v?.?.?';
        }
    }

    async loadCommands() {
        try {
            const response = await fetch('/api/commands');
            const data = await response.json();
            this.commands = data;
            this.renderCommands();
        } catch (error) {
            console.error('Failed to load commands:', error);
            this.renderFallbackCommands();
        }
    }

    updateVersionDisplay() {
        const versionElement = document.getElementById('version');
        if (this.version) {
            versionElement.textContent = `v${this.version.version}`;
            versionElement.title = `Tag: ${this.version.tag}`;
        }
    }

    renderCommands() {
        const commandsGrid = document.getElementById('commandsGrid');
        commandsGrid.innerHTML = '';

        this.commands.forEach(command => {
            const commandCard = document.createElement('div');
            commandCard.className = 'command-card';
            commandCard.innerHTML = `
                <div class="command-name">${command.name}</div>
                <div class="command-description">${command.description}</div>
                <div class="command-usage">${command.usage}</div>
            `;
            commandsGrid.appendChild(commandCard);
        });
    }

    renderFallbackCommands() {
        const fallbackCommands = [
            {
                name: 'dizz',
                description: 'Main command - shows help and available commands',
                usage: 'dizz [command] [flags]'
            },
            {
                name: 'dizz version',
                description: 'Display the current version',
                usage: 'dizz version'
            },
            {
                name: 'dizz install',
                description: 'Install dependencies and setup the project',
                usage: 'dizz install [flags]'
            }
        ];
        
        this.commands = fallbackCommands;
        this.renderCommands();
    }

    async initWasm() {
        // Check if WASM is supported
        if (!WebAssembly) {
            throw new Error('WebAssembly not supported');
        }

        // Load WASM module (this would be implemented when WASM is built)
        // For now, we'll just mark it as ready for future implementation
        this.wasmReady = true;
        console.log('WASM support detected');
    }

    async executeCommand(command) {
        if (this.wasmReady && window.runWasmCommand) {
            try {
                const result = window.runWasmCommand(command);
                return result;
            } catch (error) {
                console.error('WASM command failed:', error);
            }
        }
        
        // Fallback to API call or simulation
        return {
            success: true,
            command: command,
            output: `Simulated execution of '${command}' command`
        };
    }

    setupEventListeners() {
        // Add any global event listeners here
        document.addEventListener('keydown', (e) => {
            // Keyboard shortcuts could be implemented here
            if (e.key === 'Escape') {
                this.hideInstallCommand();
            }
        });
    }

    showInstallCommand(command) {
        const installCommand = document.getElementById('installCommand');
        installCommand.textContent = command;
        installCommand.classList.add('show');
        
        // Select the text for easy copying
        const range = document.createRange();
        range.selectNodeContents(installCommand);
        const selection = window.getSelection();
        selection.removeAllRanges();
        selection.addRange(range);
    }

    hideInstallCommand() {
        const installCommand = document.getElementById('installCommand');
        installCommand.classList.remove('show');
    }

    async checkForUpdates() {
        if (!this.version) return null;
        
        try {
            // This could check against GitHub releases API
            const response = await fetch('https://api.github.com/repos/yourusername/dizz/releases/latest');
            const latestRelease = await response.json();
            
            const currentVersion = this.version.tag;
            const latestVersion = latestRelease.tag_name;
            
            return {
                current: currentVersion,
                latest: latestVersion,
                updateAvailable: currentVersion !== latestVersion
            };
        } catch (error) {
            console.error('Failed to check for updates:', error);
            return null;
        }
    }
}

// Initialize the app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.dizzApp = new DizzApp();
});

// Export for global access
window.DizzApp = DizzApp;