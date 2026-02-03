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
                <li><a href="/docs" {{if eq .Path "/docs"}}class="active"{{end}}>Overview</a></li>
                <li><a href="#quick-start">Quick Start</a></li>
                <li><a href="#commands">Commands</a></li>
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
                <p class="tagline">Know what to work on next.</p>
                <div class="version">{{.Version}}</div>
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
        h1 { font-size: 3rem; font-weight: 700; color: #ffffff; }
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
            transition: all 0.2s ease;
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
        h2 { font-size: 2rem; margin: 40px 0 24px 0; color: #ffffff; }
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
            margin: 32px 0;
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
				.feature h3 {
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

