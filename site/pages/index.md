<h2>Quick Install</h2>

<div class="tabs">
<button class="tab active" onclick="showTab('linux')">Linux</button>
<button class="tab" onclick="showTab('macos')">macOS</button>
<button class="tab" onclick="showTab('windows')">Windows</button>
</div>

<div class="command-container">
<div id="linux" class="command-content">
<div class="command">curl -fsSL https://dizz.shitworks.co/install.sh | bash</div>
</div>
<div id="macos" class="command-content" style="display:none;">
<div class="command">curl -fsSL https://dizz.shitworks.co/install.sh | bash</div>
</div>
<div id="windows" class="command-content" style="display:none;">
<div class="command">powershell -c "irm https://dizz.shitworks.co/install.ps1 | iex"</div>
</div>
<button class="copy-btn" onclick="copyCommand()">Copy</button>
</div>

<h1>Dizz</h1>

<h2>Overview</h2>

`dizz` is a local, Git-aware developer CLI that analyzes your codebase to understand **progress**, not just correctness.  
It helps developers answer the question **“What should I work on next?”** by detecting unused code, planned work (TODOs), unstable areas, and forgotten or abandoned logic using static analysis and Git context.

Unlike linters or task managers, `dizz` models **developer intent and code evolution**. It runs fully offline, requires no configuration to start, and works across multiple languages through a unified signal-based architecture.

<h2>Features</h2>

<div class="features">

<div class="feature">
<h1>🧠</h1>
<h3>Code Intelligence</h3>
<p>Understands your codebase structure and relationships beyond simple text matching.</p>
</div>

<div class="feature">
<h1>🎯</h1>
<h3>Precision Analysis</h3>
<p>Language-aware AST parsing provides accurate insights across multiple programming languages.</p>
</div>

<div class="feature">
<h1>⚡</h1>
<h3>Millisecond Performance</h3>
<p>Enterprise-grade analysis in milliseconds, not minutes. No network calls, no tokens.</p>
</div>

<div class="feature">
<h1>📊</h1>
<h3>State Scoring</h3>
<p>Proprietary algorithm combines usage patterns, TODOs, and git history for actionable insights.</p>
</div>

<div class="feature">
<h1>🔄</h1>
<h3>Time-Aware Insights</h3>
<p>Tracks code evolution and stability to identify what's working vs what needs attention.</p>
</div>

<div class="feature">
<h1>🔒</h1>
<h3>Immutable Snapshots</h3>
<p>Content-addressed project states for perfect version control and historical analysis.</p>
</div>

</div>

<h2>CLI Commands</h2>

<div class="commands-grid">

<div class="command-card">
<div class="command-name">dizz version</div>
<div class="command-desc">Show current version and build information.</div>
<div class="command-usage">dizz version [--verbose]</div>
</div>

<div class="command-card">
<div class="command-name">dizz upgrade</div>
<div class="command-desc">Upgrade to the latest version automatically.</div>
<div class="command-usage">dizz upgrade [--force]</div>
</div>

<div class="command-card">
<div class="command-name">dizz init</div>
<div class="command-desc">Initialize dizz in your project directory.</div>
<div class="command-usage">dizz init</div>
</div>

<div class="command-card">
<div class="command-name">dizz log</div>
<div class="command-desc">Show what needs attention in your codebase.</div>
<div class="command-usage">dizz log [--all] [--verbose]</div>
</div>

<div class="command-card">
<div class="command-name">dizz status</div>
<div class="command-desc">Quick project health check with visual indicators.</div>
<div class="command-usage">dizz status</div>
</div>

<div class="command-card">
<div class="command-name">dizz snapshot</div>
<div class="command-desc">Create an immutable snapshot of project state.</div>
<div class="command-usage">dizz snapshot [--auto]</div>
</div>

<div class="command-card">
<div class="command-name">dizz list</div>
<div class="command-desc">Show all saved snapshots with timestamps and metadata.</div>
<div class="command-usage">dizz list [--format json|table]</div>
</div>

<div class="command-card">
<div class="command-name">dizz resume</div>
<div class="command-desc">Quick context after being away from the project.</div>
<div class="command-usage">dizz resume [--days N]</div>
</div>

<div class="command-card">
<div class="command-name">dizz intent</div>
<div class="command-desc">Manage intentional TODO markers and track planned work.</div>
<div class="command-usage">dizz intent add|list|remove|complete [@symbol] [description]</div>
</div>

</div>
