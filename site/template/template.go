package template

import "fmt"

func GetBaseTemplate() string {
	return BaseTemplate
}

var BaseTemplate = fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Dizz</title>
    <meta name="description" content="{{.Description}}">
    <link rel="canonical" href="https://{{.Domain}}{{.Path}}">
    <link rel="icon" href="favicon.ico" type="image/x-icon">
    <style>
			%s
		</style>
</head>
<body>
    <div class="container">
        {{if .IsDocs}}
        <div class="sidebar">
            <h2>Navigation</h2>
            <ul>
                <li><a href="/docs">Overview</a></li>
                <li><a href="#quick-start">Quick Start</a></li>
                <li><a href="#commands">Commands</a></li>
                <li><a href="#ai-agent-integration">AI Agent Integration</a></li>
                <li><a href="#symbol-states">Symbol States</a></li>
                <li><a href="#architecture">Architecture</a></li>
            </ul>
            <h2 style="margin-top: 30px;">Links</h2>
            <ul>
                <li><a href="/">← Back to Home</a></li>
                <li><a href="https://github.com/TheShiveshNetwork/dizz" target="_blank">GitHub</a></li>
            </ul>
        </div>
        {{end}}
        
        <div>
            <header>
                {{if not .IsDocs}}
                <div class="logo">
                    <img src="assets/dizz-logo.png" alt="Dizz Logo" width="128" height="128">
                </div>
                <p class="tagline">The only brain your codebase will ever need.</p>
                <div class="version">v{{.Version}}</div>
                {{end}}
            </header>

            {{if not .IsDocs}}
            <div class="nav-links">
                <a href="/">Home</a>
                <a href="/docs">Documentation</a>
                <a href="https://github.com/TheShiveshNetwork/dizz">GitHub</a>
            </div>
            {{end}}

            <main {{if .IsDocs}}class="content"{{end}}>
                {{.Content}}
            </main>

            <footer>
                <div>
                    <a href="https://github.com/TheShiveshNetwork/dizz">GitHub</a>
                    <a href="/docs">Documentation</a>
                    <a href="/">Home</a>
                </div>
                <p style="margin-top: 20px; color: #8b949e;">© 2026 Dizz. Open source and free forever.</p>
            </footer>
        </div>
    </div>

    <script>
        let currentTab = 'linux';
        
        function showTab(tab) {
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.command-content').forEach(c => c.style.display = 'none');
            
            const clickedTab = document.querySelector('[onclick="showTab(\'' + tab + '\')"]');
            if (clickedTab) {
                clickedTab.classList.add('active');
            }
            const tabContent = document.getElementById(tab);
            if (tabContent) {
                tabContent.style.display = 'block';
            }
            currentTab = tab;
        }
        
        function copyCommand() {
            const command = document.querySelector('#' + currentTab + ' .command');
            if (command) {
                navigator.clipboard.writeText(command.textContent.trim()).then(() => {
                    const btn = document.querySelector('.copy-btn');
                    if (btn) {
                        const original = btn.textContent;
                        btn.textContent = 'Copied!';
                        btn.style.background = '#10b981';
                        setTimeout(() => {
                            btn.textContent = original;
                            btn.style.background = '#58a6ff';
                        }, 2000);
                    }
                });
            }
        }

        // Scroll-based navigation highlighting
        function initScrollSpy() {
            const sections = Array.from(document.querySelectorAll('h2[id], h3[id]'));
            const allNavLinks = document.querySelectorAll('.sidebar a');
            const anchorLinks = document.querySelectorAll('.sidebar a[href^="#"]');
            const overviewLink = document.querySelector('.sidebar a[href="/docs"]');
            
            let currentActiveId = null;
            let isScrolling = false;
            
            function updateActiveNav() {
                if (isScrolling) return; // Don't update during smooth scrolling
                
                const scrollY = window.scrollY + 80; // Smaller offset for earlier activation
                let newActiveId = null;
                
                // If at the top of the page, activate Overview
                if (scrollY <= 100) {
                    newActiveId = 'overview';
                } else {
                    // Find the last section that starts before the current scroll position
                    for (let i = sections.length - 1; i >= 0; i--) {
                        const section = sections[i];
                        const sectionTop = section.offsetTop;
                        
                        if (scrollY >= sectionTop) {
                            newActiveId = section.getAttribute('id');
                            break;
                        }
                    }
                }
                
                // Only update if the active section actually changed
                if (newActiveId !== currentActiveId) {
                    // Remove active from all links
                    allNavLinks.forEach(link => link.classList.remove('active'));
                    
                    // Add active to the appropriate link
                    if (newActiveId === 'overview' && overviewLink) {
                        overviewLink.classList.add('active');
                    } else if (newActiveId) {
                        anchorLinks.forEach(link => {
                            if (link.getAttribute('href') === '#' + newActiveId) {
                                link.classList.add('active');
                            }
                        });
                    }
                    
                    currentActiveId = newActiveId;
                }
            }
            
            // Smooth scrolling for navigation links
            anchorLinks.forEach(link => {
                link.addEventListener('click', function(e) {
                    e.preventDefault();
                    const targetId = this.getAttribute('href').substring(1);
                    const targetSection = document.getElementById(targetId);
                    
                    if (targetSection) {
                        isScrolling = true;
                        const offsetTop = targetSection.offsetTop - 20;
                        
                        window.scrollTo({
                            top: offsetTop,
                            behavior: 'smooth'
                        });
                        
                        // Manually set the active state immediately on click
                        allNavLinks.forEach(l => l.classList.remove('active'));
                        this.classList.add('active');
                        currentActiveId = targetId;
                        
                        // Re-enable scroll detection after animation completes
                        setTimeout(() => {
                            isScrolling = false;
                            updateActiveNav(); // Final check to ensure correct state
                        }, 1000); // 1 second for smooth scroll to complete
                    }
                });
            });
            
            // Handle Overview link click
            if (overviewLink) {
                overviewLink.addEventListener('click', function(e) {
                    e.preventDefault();
                    isScrolling = true;
                    
                    window.scrollTo({
                        top: 0,
                        behavior: 'smooth'
                    });
                    
                    // Manually set the active state immediately on click
                    allNavLinks.forEach(l => l.classList.remove('active'));
                    this.classList.add('active');
                    currentActiveId = 'overview';
                    
                    // Re-enable scroll detection after animation completes
                    setTimeout(() => {
                        isScrolling = false;
                        updateActiveNav(); // Final check to ensure correct state
                    }, 1000);
                });
            }
            
            // Throttled scroll event listener
            let scrollTimeout;
            window.addEventListener('scroll', () => {
                clearTimeout(scrollTimeout);
                scrollTimeout = setTimeout(updateActiveNav, 50); // 50ms throttle
            });
            
            // Initial call
            setTimeout(updateActiveNav, 100); // Small delay to ensure page is rendered
        }

        // Initialize scroll spy when DOM is ready
        if (document.querySelector('.sidebar')) {
            initScrollSpy();
        }
    </script>
</body>
</html>`, BaseStyle)

const BaseStyle string = `        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, monospace; 
            line-height: 1.6; 
            background: #212122; 
            color: #e0e0e0; 
        }
        .container { 
            max-width: {{if .IsDocs}}1200px{{else}}900px{{end}}; 
            margin: 0 auto; 
            padding: 40px 20px; 
            display: flex;
            {{if .IsDocs}}gap: 40px;{{end}}
        }
        header { text-align: center; margin-bottom: 40px; {{if .IsDocs}}flex: 1;{{end}} }
        .logo { margin-bottom: 20px; color: #58a6ff; }
				h1 { margin: 30px 0; font-size: 3rem; font-weight: 700; color: #ffffff; }
        .tagline { font-size: 1.4rem; color: #8b949e; margin-bottom: 24px; }
        .version { 
            display: inline-block; 
            padding: 6px 12px; 
            background: #58a6ff; 
            color: #ffffff; 
            border-radius: 16px; 
            font-size: 0.9rem; 
            font-weight: 500;
        }
        main { 
            {{if .IsDocs}}flex: 3;{{end}}
            {{if not .IsDocs}}text-align: left;{{end}}
        }
        {{if .IsDocs}}
        .sidebar {
            flex: 1;
            min-width: 250px;
            position: sticky;
            top: 20px;
            height: fit-content;
            max-height: calc(100vh - 40px);
            overflow-y: auto;
        }
        .sidebar h2 {
            font-size: 1.2rem;
            margin-bottom: 20px;
            color: #58a6ff;
            border-bottom: 1px solid #2d2d30;
            padding-bottom: 10px;
        }
        .sidebar ul {
            list-style: none;
        }
        .sidebar li {
            margin-bottom: 8px;
        }
        .sidebar a {
            color: #8b949e;
            text-decoration: none;
            display: block;
            padding: 6px 10px;
            border-radius: 4px;
            transition: all 0.3s ease;
        }
        .sidebar a:hover {
            color: #ffffff;
            background: #2d2d30;
        }
        .sidebar a.active {
            color: #58a6ff;
            background: #1e3a5f;
        }
        .content {
            background: #1a1a1b;
            border-radius: 8px;
            padding: 30px;
            border: 1px solid #2d2d30;
        }
        {{end}}
        /* Markdown styles */
        h2 { font-size: 2rem; margin: 30px 0 24px 0; color: #ffffff; }
        h3 { font-size: 1.5rem; margin: 30px 0 15px 0; color: #ffffff; }
        h4 { font-size: 1.2rem; margin: 20px 0 10px 0; color: #ffffff; }
        p { margin-bottom: 16px; color: #e0e0e0; }
        a { color: #58a6ff; text-decoration: none; }
        a:hover { text-decoration: underline; }
        code { 
            background: #2d2d30; 
            padding: 2px 6px; 
            border-radius: 3px; 
            font-family: 'Monaco', 'Menlo', monospace; 
            font-size: 0.9em; 
            color: #f8f8f2;
        }
        pre {
            background: #1a1a1b;
            border-radius: 6px;
            padding: 16px;
            overflow-x: auto;
            margin: 20px 0;
            border: 1px solid #2d2d30;
        }
        pre code {
            background: none;
            padding: 0;
            color: #e0e0e0;
        }
        blockquote {
            border-left: 4px solid #58a6ff;
            padding-left: 20px;
            margin: 24px 0;
            color: #8b949e;
            font-style: italic;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        th, td {
            border: 1px solid #2d2d30;
            padding: 12px;
            text-align: left;
        }
        th {
            background: #2d2d30;
            color: #ffffff;
            font-weight: 600;
        }
        ul, ol {
            margin: 16px 0;
            padding-left: 30px;
        }
        li {
            margin-bottom: 8px;
            color: #e0e0e0;
        }
        .command-box {
            background: #2d2d30;
            border-radius: 6px;
            padding: 20px;
            margin: 20px 0;
            border: 1px solid #58a6ff;
        }
        .command-title {
            color: #58a6ff;
            font-weight: 600;
            margin-bottom: 8px;
        }
        .command-desc {
            color: #8b949e;
            margin-bottom: 12px;
        }
        .command-usage {
            background: #1a1a1b;
            padding: 8px 12px;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.9rem;
            color: #9ca3af;
        }
        /* Tab styles for quick install */
        .tabs {
            display: flex;
            gap: 2px;
            margin-bottom: 0;
        }
        .tab {
            padding: 12px 20px;
            background: #2d2d30;
            border: none;
            color: #8b949e;
            cursor: pointer;
            font-family: inherit;
            font-size: 0.9rem;
            transition: all 0.2s ease;
        }
        .tab:hover {
            background: #3d3d40;
        }
        .tab.active {
            background: #58a6ff;
            color: #ffffff;
        }
        .command-container {
            background: #1a1a1b;
            border-radius: 6px;
            padding: 20px;
            position: relative;
        }
        .command {
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.9rem;
            color: #9ca3af;
            margin-bottom: 12px;
            word-break: break-all;
            padding-right: 80px;
        }
        .copy-btn {
            position: absolute;
            top: 12px;
            right: 12px;
            padding: 6px 12px;
            background: #58a6ff;
            color: #ffffff;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.8rem;
            transition: background-color 0.2s ease;
        }
        .copy-btn:hover {
            background: #4692e6;
        }
        .features {
            display: grid;
            gap: 24px;
            margin-bottom: 32px;
        }
        .feature {
            padding: 16px;
            background: #2d2d30;
            border-radius: 6px;
        }
				.feature h1, .feature h3 {
					margin: 0;
				}
        .commands-grid {
            display: grid;
            gap: 16px;
            margin-bottom: 32px;
        }
        .command-card {
            padding: 20px;
            background: #2d2d30;
            border-radius: 6px;
        }
        footer { text-align: center; padding: 40px 0; border-top: 1px solid #2d2d30; }
        footer a { color: #58a6ff; text-decoration: none; margin: 0 16px; }
        footer a:hover { text-decoration: underline; }
        .nav-links {
            text-align: center;
            margin-bottom: 30px;
        }
        .nav-links a {
            color: #58a6ff;
            text-decoration: none;
            margin: 0 10px;
            padding: 8px 16px;
            border-radius: 4px;
            transition: background-color 0.2s ease;
        }
        .nav-links a:hover {
            background: #2d2d30;
        }
        /* Ensure proper centering and spacing */
        @media (min-width: 769px) {
            .features {
                grid-template-columns: repeat(3, 1fr);
            }
        }
        
        @media (max-width: 768px) { 
            .container { 
                padding: 20px 16px; 
                flex-direction: column;
            }
            h1 { font-size: 2.5rem; }
            .sidebar {
                position: static;
                max-height: none;
                order: 2;
            }
            main {
                order: 1;
            }
            .features {
                grid-template-columns: 1fr;
            }
        }
    `
