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

<blockquote>
<strong>Important:</strong> After installing, run <code>dizz install-skill</code> to enable AI agent discovery. Agents like Claude Code, Cursor, and Gemini CLI can then use dizz automatically.
</blockquote>

<h1>Dizz</h1>

<h2>Overview</h2>

`dizz` reads your Git history, looks at your code, and tells you what needs doing. TODOs, half-finished work, things changing too often — it tracks all of it. Takes snapshots so you never lose context. Agents love it because it speaks their language.

No setup. No internet. Won't touch your code.

<h2>Features</h2>

<div class="features">

<div class="feature">
<h1>🧠</h1>
<h3>Code Intelligence</h3>
<p>Reads your Git history and analyzes your code to surface what needs attention. No configuration needed.</p>
</div>

<div class="feature">
<h1>🤖</h1>
<h3>AI Agent Ready</h3>
<p>Speaks TON — a token-optimized format AI agents understand instantly. Auto-discovers via <code>dizz install-skill</code>.</p>
</div>

<div class="feature">
<h1>⚡</h1>
<h3>Millisecond Performance</h3>
<p>Enterprise-grade analysis in milliseconds, not minutes. No network calls, no tokens, no cloud.</p>
</div>

<div class="feature">
<h1>📊</h1>
<h3>State Tracking</h3>
<p>Tracks every symbol as active, planned, unstable, unused, or abandoned — based on real usage and Git history.</p>
</div>

<div class="feature">
<h1>🔒</h1>
<h3>Immutable Snapshots</h3>
<p>Content-addressed project states for perfect version control and historical analysis.</p>
</div>

<div class="feature">
<h1>🎯</h1>
<h3>Intent System</h3>
<p>Long-lived project intents separate from throwaway TODOs. Never lose track of why you wrote something.</p>
</div>

</div>

<h2>CLI Commands</h2>

<div class="commands-grid">

<div class="command-card">
<div class="command-name">dizz init</div>
<div class="command-desc">Initialize dizz in your project directory.</div>
<div class="command-usage">dizz init</div>
</div>

<div class="command-card">
<div class="command-name">dizz log</div>
<div class="command-desc">Full analysis — planned work, unstable areas, unused and abandoned code.</div>
<div class="command-usage">dizz log [--all] [--verbose]</div>
</div>

<div class="command-card">
<div class="command-name">dizz status</div>
<div class="command-desc">Quick project health snapshot with visual indicators.</div>
<div class="command-usage">dizz status</div>
</div>

<div class="command-card">
<div class="command-name">dizz context</div>
<div class="command-desc">Token-optimized project dump for AI agents (~2 KB). Pipe-delimited, no parser needed.</div>
<div class="command-usage">dizz context [--intents] [--symbols] [--todos]</div>
</div>

<div class="command-card">
<div class="command-name">dizz snapshot</div>
<div class="command-desc">Create immutable, content-addressed state records with delta support.</div>
<div class="command-usage">dizz snapshot [--auto] [--diff]</div>
</div>

<div class="command-card">
<div class="command-name">dizz intent</div>
<div class="command-desc">Track long-lived project goals separate from throwaway TODOs.</div>
<div class="command-usage">dizz intent add|list|resolve</div>
</div>

<div class="command-card">
<div class="command-name">dizz install-skill</div>
<div class="command-desc">Auto-detect AI agents and install the dizz skill for automatic discovery.</div>
<div class="command-usage">dizz install-skill</div>
</div>

<div class="command-card">
<div class="command-name">dizz list</div>
<div class="command-desc">Show all saved snapshots with timestamps and metadata.</div>
<div class="command-usage">dizz list</div>
</div>

<div class="command-card">
<div class="command-name">dizz resume</div>
<div class="command-desc">Quick context after being away from the project.</div>
<div class="command-usage">dizz resume</div>
</div>

<div class="command-card">
<div class="command-name">dizz version</div>
<div class="command-desc">Show current version and build information.</div>
<div class="command-usage">dizz version [--verbose]</div>
</div>

<div class="command-card">
<div class="command-name">dizz upgrade</div>
<div class="command-desc">Upgrade to the latest version automatically.</div>
<div class="command-usage">dizz upgrade</div>
</div>

</div>
