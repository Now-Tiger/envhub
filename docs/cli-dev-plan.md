# EnvHub CLI - Production Implementation Plan

## Executive Summary

Transform the basic `flag`-based CLI into a professional, visually stunning developer tool. This plan prioritizes **lightning-fast startup** (<50ms), **delightful developer experience**, and **battle-tested reliability**.

---

## 1. CLI Framework Selection

### Recommendation: **Cobra + Bubble Tea Hybrid**

| Approach | Pros | Cons | Recommendation |
|----------|------|------|----------------|
| **Cobra** | Industry standard (Docker, Kubernetes), auto-generated help, subcommands, flags | No built-in UI | **Primary** - for command structure |
| **Bubble Tea** | Amazing TUI animations, spinners, interactive inputs | Higher overhead, complexity | **Secondary** - for interactive prompts |
| **urfave/cli v2** | Simpler than Cobra, good flag parsing | Less adoption in K8s ecosystem | Consider as alternative |
| **charm.sh libraries** | Modern, beautiful, lightweight | Newer ecosystem | Great for styling (lipgloss) |

### Selected Stack

```go
// go.mod additions
github.com/spf13/cobra@v1.9.1          // Command framework
github.com/charmbracelet/lipgloss@v1.0.0 // Terminal styling
github.com/charmbracelet/bubbles@v0.20.0 // TUI components (spinner, progress)
github.com/charmbracelet/bubbly@v0.1.0   // Progress bars
github.com/mitchellh/go-homedir@v1.1.0  // Config paths
github.com/spf13/viper@v1.19.0          // Config management
github.com/AlecAivazis/survey/v2@v2.0.9 // Interactive prompts
```

### Justification

- **Cobra**: Provides battle-tested subcommand hierarchy, flag validation, and help generation
- **lipgloss**: Lightweight styling (not full TUI), zero runtime overhead for static output
- **bubbles**: On-demand loading for spinners/progress (lazy import)
- **No full Bubble Tea TUI**: Avoid complexity for non-interactive commands

---

## 2. ASCII Art & Animation Strategy

### Approach: **Hybrid - Pre-rendered Static + Runtime Animated**

#### A. Static Banner (Every Invocation)

```go
// internal/cli/banner/banner.go
package banner

var EnvHubBanner = `
██████╗ ██╗██████╗ ███████╗██╗     ██╗███╗   ██╗███████╗
██╔══██╗██║██╔══██╗██╔════╝██║     ██║████╗  ██║██╔════╝
██████╔╝██║██████╔╝█████╗  ██║     ██║██╔██╗ ██║█████╗  
██╔══██╗██║██╔══██╗██╔══╝  ██║     ██║██║╚██╗██║██╔══╝  
██║  ██║██║██║  ██║███████╗███████╗██║██║ ╚████║███████╗
╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚══════╝╚══════╝╚═╝╚═╝  ╚═══╝╚══════╝
                                                   `
```

#### B. Runtime Animation (Loading States)

```go
// internal/cli/spinner/spinner.go
package spinner

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/charmbracelet/bubbles/spinner"
)

type Spinner struct {
    s    spinner.Model
    styl lipgloss.Style
}

func New() Spinner {
    s, _ := spinner.New(
        spinner.WithSpinner(spinner.Line),
        spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("86"))),
    )
    return Spinner{s: s}
}
```

#### C. Strategy Decision Matrix

| When | What | Why |
|------|------|-----|
| `--help` | Full ASCII banner + usage | First impression, marketing |
| Every command | Compact 1-line header | Brand reinforcement |
| Long operations | Spinner | User feedback, perceived performance |
| Errors | Red highlights + suggestions | Clarity, actionability |
| First run | Animated welcome + setup wizard | Onboarding |

---

## 3. Command Structure

### Proposed Hierarchy

```
envhub [global flags]
├── login              # Interactive login flow
│   └── --browser     # Open browser for web login
├── pull              # Pull secrets
│   ├── --project     # Project name (required)
│   ├── --env        # Environment (default: development)
│   ├── --output     # Output file (.env)
│   └── --format     # Output format: env, json, yaml, docker
├── push              # Push secrets (NEW)
├── list              # List projects
│   ├── --org        # Filter by organization
│   └── --format    # Table, JSON, YAML
├── whoami            # Current user info
├── org              # Organization management
│   ├── list         # List organizations
│   ├── members      # List org members
│   └── invite       # Invite member
├── team              # Team management (NEW)
│   ├── list
│   ├── add
│   └── remove
├── init              # Initialize in current directory
├── logout            # Clear cached credentials
├── config            # Manage configuration
│   ├── show         # Show current config
│   ├── set          # Set config value
│   └── edit         # Open config in editor
├── update           # Check for updates
└── version           # Show version info
```

### Global Flags

```go
var (
    verbose   bool   // -v, --verbose
    config    string // --config path
    output    string // -o, --output format
    no-color  bool   // --no-color (accessibility)
)
```

---

## 4. File Structure

```
cmd/cli/
├── main.go                          # Entry point
└── cmd/
    ├── root.go                      # Root command
    ├── login.go                     # Login command
    ├── pull.go                      # Pull secrets
    ├── push.go                      # Push secrets  
    ├── list.go                      # List projects
    ├── whoami.go                    # Who am I
    ├── org/
    │   ├── org.go                   # Org parent
    │   ├── list.go
    │   ├── members.go
    │   └── invite.go
    ├── team/
    │   ├── team.go
    │   ├── list.go
    │   ├── add.go
    │   └── remove.go
    ├── init.go                      # Project init
    ├── logout.go
    ├── config.go
    ├── update.go
    └── version.go

internal/
├── cli/
│   ├── banner/
│   │   └── banner.go               # ASCII art
│   ├── spinner/
│   │   └── spinner.go              # Loading indicators
│   ├── style/
│   │   └── style.go                # Lipgloss styles
│   ├── prompt/
│   │   └── prompt.go               # Interactive prompts
│   └── progress/
│       └── progress.go             # Progress bars
├── config/
│   ├── config.go                   # Viper config
│   └── auth.go                     # Token management
├── client/
│   └── api.go                      # HTTP client wrapper
└── cache/
    └── cache.go                    # Response caching

pkg/
├── types/
│   └── cli.go                      # Existing CLI types
└── crypto/
    └── tokens.go                   # Token encryption (reuse existing)

scripts/
├── install.sh                      # Installation script
└── completions/
    ├── bash.sh                     # Bash completions
    ├── zsh.sh                      # Zsh completions
    └── fish.sh                     # Fish completions
```

---

## 5. Performance & Efficiency

### A. Lazy Loading Strategy

```go
// internal/cli/banner/banner.go
// Only load heavy ASCII art when needed

import _ "embed"  // Go 1.16+ embed

//go:embed banner.txt
var bannerStatic string

// GetBanner returns pre-compiled banner (zero allocation after init)
func GetBanner() string {
    return bannerStatic
}
```

### B. Config Caching

```go
// internal/config/config.go
type Config struct {
    APIBaseURL string `mapstructure:"api_base_url"`
    Token      string `mapstructure:"-"` // Never cache raw token
    TokenHash  string `mapstructure:"token_hash"` // Encrypted
    CacheTTL   time.Duration
    
    mu sync.RWMutex
}
```

### C. Token Storage Security

- Use existing crypto package - AES-256-GCM
- Store in macOS Keychain, Linux libsecret, Windows Credential Manager

### D. Parallel Operations

```go
// For bulk operations (pull multiple environments)
func (p *PullCmd) Run(args []string) error {
    var wg sync.WaitGroup
    results := make(chan SecretResult, len(envs))
    
    for _, env := range envs {
        wg.Add(1)
        go func(e string) {
            defer wg.Done()
            results <- p.fetchSecrets(project, e)
        }(env)
    }
    // Collect and merge results
}
```

---

## 6. User Experience

### A. Interactive Prompts (Login Flow)

```go
// cmd/cli/login.go
var loginSurvey = []*survey.Question{
    {
        Name: "email",
        Prompt: &survey.Input{
            Message: "Email:",
            Help:    "Your EnvHub account email",
        },
        Validate: survey.Required,
    },
    {
        Name: "password",
        Prompt: &survey.Password{
            Message: "Password:",
        },
        Validate: survey.Required,
    },
}
```

### B. Colored Output (lipgloss)

```go
// internal/cli/style/style.go
var (
    Header = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("86"))
    
    Success = lipgloss.NewStyle().
        Foreground(lipgloss.Color("82"))
    
    Error = lipgloss.NewStyle().
        Foreground(lipgloss.Color("196"))
)
```

### C. Error Handling with Suggestions

```go
// Instead of: "error: token required"
// Show: 
//   ✗ Authentication failed
//   Hint: Set your API token:
//     • Export: export ENVUB_TOKEN=your_token
//     • Login:  envhub login
```

---

## 7. Implementation Phases

### Phase 1: Foundation (Week 1)
- [x] Replace `flag` with Cobra framework
- [ ] Set up project structure (cmd/cli/cmd/*)
- [ ] Implement global flags (--config, --verbose, --no-color)
- [ ] Basic command routing (login, pull, list, whoami)

### Phase 2: Visual Branding (Week 1-2)
- [ ] Add ASCII banner with lipgloss styling
- [ ] Implement spinner for loading states
- [ ] Color-coded output (success/error/warning)
- [ ] Compact header on every command

### Phase 3: User Experience (Week 2)
- [ ] Interactive login with survey
- [ ] Config management (viper)
- [ ] Token encryption (reuse existing crypto)
- [ ] Help improvements with examples

### Phase 4: Feature Complete (Week 3)
- [ ] All subcommands (org, team, init, config)
- [ ] Pull multiple environments in parallel
- [ ] Push secrets command
- [ ] Output format options (env, json, yaml, docker)

### Phase 5: Polish & Distribution (Week 4)
- [ ] Shell completions (bash, zsh, fish)
- [ ] Auto-update mechanism
- [ ] Installation scripts (brew, go install)
- [ ] Documentation

---

## 8. Key Design Decisions

| Decision | Tradeoff | Rationale |
|----------|----------|-----------|
| **Cobra over urfav** | Slightly more boilerplate | Industry standard, Kubernetes/Docker adoption |
| **Static ASCII banner** | Slightly larger binary (~5KB) | Zero runtime generation cost |
| **On-demand spinners** | Slight delay on first load | Faster startup for simple commands |
| **lipgloss over full TUI** | Limited interactivity | 10x faster, sufficient for CLI |
| **Survey for prompts** | External dependency | Battle-tested, accessible |
| **File-based config** | Less secure than OS keyring | Cross-platform simplicity |

---

## 9. Distribution Plan

### Installation Methods

```bash
# 1. Homebrew (recommended for macOS)
brew install envhub/tap/envhub

# 2. Go install
go install github.com/Now-Tiger/envhub/cmd/cli@latest

# 3. Direct binary
curl -sL https://envhub.dev/install.sh | sh

# 4. Docker
docker run --rm envhub/cli pull -p myapp -e prod
```

---

## Summary

This plan transforms the basic CLI into a **professional, developer-friendly tool** that:

1. **Starts lightning fast** - Lazy loading, pre-compiled assets
2. **Looks stunning** - ASCII branding, colored output, spinners
3. **Works great** - Interactive prompts, helpful errors, parallel ops
4. **Distributes easily** - Multiple install methods, auto-updates

The implementation prioritizes **incremental delivery**: Phase 1 gives you a working Cobra-based CLI in week 1, with visual polish added progressively.
