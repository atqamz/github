package main

// spec is one managed repository. Everything the baseline in apply() already
// sets is deliberately absent here, so a row carries only what is unique to
// that repo. Zero values mean: public, issue tracker on, no homepage, no
// topics, no pre-existing ruleset.
type spec struct {
	Name        string
	Description string
	Homepage    string
	Topics      []string
	Private     bool
	NoIssues    bool

	// RequirePR additionally bars direct pushes to the default branch, so
	// changes have to arrive as a squash-merged pull request.
	RequirePR bool

	// CanApprovePRs lets Actions open and approve pull requests, which only a
	// repo running a release-bot workflow needs.
	CanApprovePRs bool

	// NoDependabotSecurityUpdates turns off automated security fix PRs, for a
	// repo where dependabot's fetcher cannot work.
	NoDependabotSecurityUpdates bool

	// RulesetID is the numeric id of a branch ruleset that already exists on
	// GitHub. Set it so the adopt run takes that ruleset over instead of
	// creating a second one alongside it.
	RulesetID string
}

var repos = []spec{
	{
		Name:        "atqamz",
		Description: "Config files for my GitHub profile.",
		Homepage:    "https://github.com/atqamz",
		Topics:      []string{"config", "github-config"},
		NoIssues:    true,
	},
	{
		Name:        "atqamz.github.io",
		Description: "personal dump",
		Homepage:    "https://atqamz.com",
	},
	{
		Name:        "dotagents",
		Description: "Model-agnostic operating config for AI coding agents (Claude Code, opencode): canonical AGENTS.md rules, skills, hooks",
	},
	{
		Name:        "dotfiles",
		Description: "Out-of-store dotfiles (Hyprland, caelestia, Zed, ...) live-symlinked from my NixOS config",
	},
	{
		Name:        "github",
		Description: "Personal GitHub repo config-as-code (Pulumi + Go)",
	},
	{
		Name:        "hawa",
		Description: "Privacy-first period tracking app for Indonesian women",
		Private:     true,
	},
	{
		Name: "mitemiru",
	},
	{
		Name: "omanixy",
	},
	{
		Name:    "password-store",
		Private: true,
	},
	{
		Name:        "rucika",
		Description: "Portable isolated personal research agent for Telegram/WhatsApp",
		Private:     true,
	},
	{
		Name:          "secondhand",
		Description:   "Talk to one agent. Ship with a crew. CLI: hand",
		RequirePR:     true,
		CanApprovePRs: true,
		RulesetID:     "19698130",
	},
	{
		Name:        "secondpeer",
		Description: "You code. It watches, remembers, and helps. CLI: peer",
	},
	{
		Name:                        "universe",
		Description:                 "Personal NixOS flake: Hyprland + caelestia desktop across two laptops, with home-manager, sops-nix, and disko",
		NoDependabotSecurityUpdates: true,
	},
	{
		Name:        "vault",
		Description: "private encrypted credentials companion to atqamz/universe",
		Private:     true,
	},
	{
		Name:        "wp",
		Description: "Mobile-first wedding and household planner PWA on Cloudflare Workers",
	},
	{
		Name:        "zen-profile",
		Description: "Zen browser profile sync (age-encrypted session blobs)",
		Private:     true,
	},
}
