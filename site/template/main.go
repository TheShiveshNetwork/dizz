package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

type Feature struct {
	Icon        string
	Title       string
	Description string
}

type Command struct {
	Name        string
	Description string
	Usage       string
	Flags       []string
}

type SiteData struct {
	Version  string
	Domain   string
	Features []Feature
	Commands []Command
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dizz - Progress-aware Dev CLI Tool</title>
    <meta name="description" content="Dizz is a progress-aware CLI tool that analyzes your codebase to detect unused functions, TODOs, and project state. Get intelligent development insights and never waste time deciding what to work on next.">
    <meta name="keywords" content="CLI tool, code analysis, dead code detection, TODO tracking, development workflow, git integration, project health">
    <meta name="author" content="Dizz">
    <meta property="og:title" content="Dizz - Progress-aware Dev CLI Tool">
    <meta property="og:description" content="Know what to work on next. Analyze your codebase for unused functions, TODOs, and project insights.">
    <meta property="og:type" content="website">
    <meta property="og:url" content="https://{{.Domain}}">
    <meta property="og:image" content="https://{{.Domain}}/assets/dizz-logo.png">
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:title" content="Dizz - Progress-aware Dev CLI Tool">
    <meta name="twitter:description" content="Know what to work on next. Analyze your codebase for unused functions, TODOs, and project insights.">
    <meta name="twitter:image" content="https://{{.Domain}}/assets/dizz-logo.png">
    <link rel="canonical" href="https://{{.Domain}}">
    <link rel="icon" href="favicon.ico" type="image/x-icon">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, monospace; 
            line-height: 1.6; 
            background: #212122; 
            color: #e0e0e0; 
        }
        .container { max-width: 900px; margin: 0 auto; padding: 40px 20px; }
        header { text-align: center; margin-bottom: 60px; }
        .logo { margin-bottom: 20px; color: #58a6ff; }
        h1 { font-size: 3rem; font-weight: 700; margin-bottom: 16px; color: #ffffff; }
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
        section { margin-bottom: 60px; }
        h2 { font-size: 2rem; margin-bottom: 24px; color: #ffffff; }
        .tabs { display: flex; gap: 2px; margin-bottom: 0; }
        .tab { 
            padding: 12px 20px; 
            background: #2d2d30; 
            border: none; 
            color: #8b949e; 
            cursor: pointer; 
            font-family: inherit; 
            font-size: 0.9rem;
        }
        .tab.active { background: #212122; color: #ffffff; }
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
        }
        .features { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 24px; }
        .feature { padding: 24px; background: #2d2d30; border-radius: 6px; }
        .feature h3 { margin-bottom: 12px; color: #58a6ff; }
        .commands-grid { display: grid; gap: 16px; }
        .command-card { padding: 20px; background: #2d2d30; border-radius: 6px; }
        .command-name { font-weight: 600; color: #58a6ff; margin-bottom: 8px; }
        .command-desc { color: #8b949e; margin-bottom: 12px; }
        .command-usage { 
            font-family: 'Monaco', 'Menlo', monospace; 
            background: #1a1a1b; 
            padding: 8px 12px; 
            border-radius: 4px; 
            font-size: 0.85rem; 
            color: #9ca3af;
        }
        footer { text-align: center; padding: 40px 0; border-top: 1px solid #2d2d30; }
        footer a { color: #58a6ff; text-decoration: none; margin: 0 16px; }
        footer a:hover { text-decoration: underline; }
        @media (max-width: 768px) { 
            .container { padding: 20px 16px; }
            h1 { font-size: 2.5rem; }
            .features { grid-template-columns: 1fr; }
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <div class="logo">
                <img src="assets/dizz-logo.png" alt="Dizz Logo" width="128" height="128">
            </div>
            <p class="tagline">Know what to work on next.</p>
            <div class="version">{{.Version}}</div>
        </header>

				<h1>Dizz</h1>
        <section>
            <h2>Quick Install</h2>
            <div class="tabs">
                <button class="tab active" onclick="showTab('linux')">Linux</button>
                <button class="tab" onclick="showTab('macos')">macOS</button>
                <button class="tab" onclick="showTab('windows')">Windows</button>
            </div>
            <div class="command-container">
                <div id="linux" class="command-content">
                    <div class="command">curl -sSL https://{{.Domain}}/install.sh | bash</div>
                </div>
                <div id="macos" class="command-content" style="display:none;">
                    <div class="command">curl -sSL https://{{.Domain}}/install.sh | bash</div>
                </div>
                <div id="windows" class="command-content" style="display:none;">
                    <div class="command">powershell -c "irm https://{{.Domain}}/install.ps1 | iex"</div>
                </div>
                <button class="copy-btn" onclick="copyCommand()">Copy</button>
            </div>
        </section>

        <section>
            <h2>Features</h2>
            <div class="features">
                {{range .Features}}
                <div class="feature">
                    <h3>{{.Icon}} {{.Title}}</h3>
                    <p>{{.Description}}</p>
                </div>
                {{end}}
            </div>
        </section>

        <section>
            <h2>CLI Commands</h2>
            <div class="commands-grid">
                {{range .Commands}}
                <div class="command-card">
                    <div class="command-name">{{.Name}}</div>
                    <div class="command-desc">{{.Description}}</div>
                    <div class="command-usage">{{.Usage}}</div>
                </div>
                {{end}}
            </div>
        </section>

        <footer>
            <div>
                <a href="https://github.com/TheShiveshNetwork/dizz">GitHub</a>
                <a href="/docs">Documentation</a>
                <a href="/api/version">API</a>
            </div>
            <p style="margin-top: 20px; color: #8b949e;">© 2024 Dizz. Open source and free forever.</p>
        </footer>
    </div>

    <script>
        let currentTab = 'linux';
        
        function showTab(tab) {
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.command-content').forEach(c => c.style.display = 'none');
            
            document.querySelector('[onclick="showTab(\'' + tab + '\')"]').classList.add('active');
            document.getElementById(tab).style.display = 'block';
            currentTab = tab;
        }
        
        function copyCommand() {
            const command = document.querySelector('#' + currentTab + ' .command').textContent.trim();
            navigator.clipboard.writeText(command).then(() => {
                const btn = document.querySelector('.copy-btn');
                const original = btn.textContent;
                btn.textContent = 'Copied!';
                btn.style.background = '#10b981';
                setTimeout(() => {
                    btn.textContent = original;
                    btn.style.background = '#58a6ff';
                }, 2000);
            });
        }
    </script>
</body>
</html>`

func getSiteData() SiteData {
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "dizz.shitworks.co"
	}

	version := os.Getenv("VERSION")
	if version == "" {
		version = "v0.0.0"
	}

	features := []Feature{
		{
			Icon:        "🧠",
			Title:       "Code Intelligence",
			Description: "Understands your codebase structure and relationships beyond simple text matching.",
		},
		{
			Icon:        "🎯",
			Title:       "Precision Analysis",
			Description: "Language-aware AST parsing provides accurate insights across multiple programming languages.",
		},
		{
			Icon:        "⚡",
			Title:       "Millisecond Performance",
			Description: "Enterprise-grade analysis in milliseconds, not minutes. No network calls, no tokens.",
		},
		{
			Icon:        "📊",
			Title:       "State Scoring",
			Description: "Proprietary algorithm combines usage patterns, TODOs, and git history for actionable insights.",
		},
		{
			Icon:        "🔄",
			Title:       "Time-Aware Insights",
			Description: "Tracks code evolution and stability to identify what's working vs what needs attention.",
		},
		{
			Icon:        "🔒",
			Title:       "Immutable Snapshots",
			Description: "Content-addressed project states for perfect version control and historical analysis.",
		},
	}

	commands := []Command{
		{
			Name:        "dizz version",
			Description: "Show current version and build information.",
			Usage:       "dizz version [--verbose]",
		},
		{
			Name:        "dizz upgrade",
			Description: "Upgrade to the latest version automatically.",
			Usage:       "dizz upgrade [--force]",
		},
		{
			Name:        "dizz init",
			Description: "Initialize dizz in your project directory.",
			Usage:       "dizz init",
		},
		{
			Name:        "dizz log",
			Description: "Show what needs attention in your codebase.",
			Usage:       "dizz log [--all] [--verbose]",
		},
		{
			Name:        "dizz status",
			Description: "Quick project health check with visual indicators.",
			Usage:       "dizz status",
		},
		{
			Name:        "dizz snapshot",
			Description: "Create an immutable snapshot of project state.",
			Usage:       "dizz snapshot [--auto]",
		},
		{
			Name:        "dizz list",
			Description: "Show all saved snapshots with timestamps and metadata.",
			Usage:       "dizz list [--format json|table]",
		},
		{
			Name:        "dizz resume",
			Description: "Quick context after being away from the project.",
			Usage:       "dizz resume [--days N]",
		},
		{
			Name:        "dizz intent",
			Description: "Manage intentional TODO markers and track planned work.",
			Usage:       "dizz intent add|list|remove|complete [@symbol] [description]",
		},
	}

	return SiteData{
		Version:  version,
		Domain:   domain,
		Features: features,
		Commands: commands,
	}
}

func main() {
	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		os.Exit(1)
	}

	data := getSiteData()

	var output strings.Builder
	if err := tmpl.Execute(&output, data); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		os.Exit(1)
	}

	// Write to stdout or file
	outputFile := "index.html"
	if len(os.Args) > 1 {
		outputFile = os.Args[1]
	}

	if err := os.WriteFile(outputFile, []byte(output.String()), 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s successfully\n", outputFile)
}
