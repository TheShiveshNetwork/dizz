// Install script handling
class InstallManager {
    constructor() {
        this.os = this.detectOS();
        this.installCommands = {
            unix: 'curl -fsSL https://your-domain.com/install.sh | sh',
            windows: 'iwr -useb https://your-domain.com/install.ps1 | iex'
        };
    }

    detectOS() {
        const userAgent = navigator.userAgent;
        const platform = navigator.platform;
        
        if (platform.indexOf('Win') !== -1 || userAgent.indexOf('Windows') !== -1) {
            return 'windows';
        } else if (platform.indexOf('Mac') !== -1 || userAgent.indexOf('Mac') !== -1) {
            return 'macos';
        } else if (platform.indexOf('Linux') !== -1 || userAgent.indexOf('Linux') !== -1) {
            return 'linux';
        } else {
            return 'unknown';
        }
    }

    getInstallCommand(os) {
        if (os === 'windows') {
            return this.installCommands.windows;
        } else {
            return this.installCommands.unix;
        }
    }

    copyInstallScript(os) {
        const command = this.getInstallCommand(os);
        
        if (navigator.clipboard) {
            navigator.clipboard.writeText(command).then(() => {
                this.showInstallCommand(command, os);
                this.showCopySuccess();
            }).catch(err => {
                console.error('Failed to copy text: ', err);
                this.fallbackCopy(command, os);
            });
        } else {
            this.fallbackCopy(command, os);
        }
    }

    fallbackCopy(command, os) {
        const textArea = document.createElement('textarea');
        textArea.value = command;
        textArea.style.position = 'fixed';
        textArea.style.left = '-999999px';
        textArea.style.top = '-999999px';
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();
        
        try {
            document.execCommand('copy');
            this.showInstallCommand(command, os);
            this.showCopySuccess();
        } catch (err) {
            console.error('Fallback copy failed: ', err);
            this.showCopyError();
        }
        
        document.body.removeChild(textArea);
    }

    showInstallCommand(command, os) {
        const installCommand = document.getElementById('installCommand');
        const osName = os === 'windows' ? 'Windows' : 'Unix/macOS';
        
        installCommand.innerHTML = `
            <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>${command}</span>
                <button onclick="installManager.copyInstallScript('${os}')" 
                        style="background: none; border: 1px solid #64748b; color: #94a3b8; 
                               padding: 4px 8px; border-radius: 4px; cursor: pointer; font-size: 0.8rem;">
                    Copy
                </button>
            </div>
            <div style="font-size: 0.8rem; margin-top: 8px; color: #64748b;">
                Run this command in your terminal to install Dizz on ${osName}
            </div>
        `;
        installCommand.classList.add('show');
    }

    showCopySuccess() {
        this.showToast('Copied to clipboard!', 'success');
    }

    showCopyError() {
        this.showToast('Failed to copy. Please copy manually.', 'error');
    }

    showToast(message, type = 'info') {
        // Remove existing toasts
        const existingToast = document.querySelector('.toast');
        if (existingToast) {
            existingToast.remove();
        }

        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;
        
        // Style the toast
        toast.style.cssText = `
            position: fixed;
            bottom: 20px;
            right: 20px;
            padding: 12px 20px;
            border-radius: 8px;
            color: white;
            font-weight: 500;
            z-index: 1000;
            animation: slideIn 0.3s ease;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        `;
        
        if (type === 'success') {
            toast.style.background = '#10b981';
        } else if (type === 'error') {
            toast.style.background = '#ef4444';
        } else {
            toast.style.background = '#3b82f6';
        }
        
        document.body.appendChild(toast);
        
        // Auto remove after 3 seconds
        setTimeout(() => {
            toast.style.animation = 'slideOut 0.3s ease';
            setTimeout(() => {
                if (toast.parentNode) {
                    toast.parentNode.removeChild(toast);
                }
            }, 300);
        }, 3000);
    }

    getInstallationInstructions() {
        const os = this.os;
        let instructions = '';
        
        switch (os) {
            case 'windows':
                instructions = `
                    <h3>Windows Installation</h3>
                    <ol>
                        <li>Open PowerShell as Administrator</li>
                        <li>Run: ${this.installCommands.windows}</li>
                        <li>Follow the installation prompts</li>
                        <li>Restart your terminal</li>
                    </ol>
                `;
                break;
            case 'macos':
                instructions = `
                    <h3>macOS Installation</h3>
                    <ol>
                        <li>Open Terminal</li>
                        <li>Run: ${this.installCommands.unix}</li>
                        <li>Enter your password when prompted</li>
                        <li>The installation will complete automatically</li>
                    </ol>
                `;
                break;
            case 'linux':
                instructions = `
                    <h3>Linux Installation</h3>
                    <ol>
                        <li>Open your terminal</li>
                        <li>Run: ${this.installCommands.unix}</li>
                        <li>Enter your password when prompted</li>
                        <li>The installation will complete automatically</li>
                    </ol>
                `;
                break;
            default:
                instructions = `
                    <h3>Manual Installation</h3>
                    <p>Please visit our <a href="https://github.com/yourusername/dizz/releases">GitHub releases</a> 
                    to download the appropriate binary for your operating system.</p>
                `;
        }
        
        return instructions;
    }
}

// Add CSS animations for toasts
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
    
    @keyframes slideOut {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(100%);
            opacity: 0;
        }
    }
`;
document.head.appendChild(style);

// Make available globally
window.installManager = new InstallManager();

// Global function for button onclick handlers
window.copyInstallScript = function(os) {
    window.installManager.copyInstallScript(os);
};