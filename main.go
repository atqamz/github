package main

import (
	"slices"
	"strings"

	"github.com/pulumi/pulumi-github/sdk/v6/go/github"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		adopt := strings.Split(config.Get(ctx, "adopt"), ",")

		for _, s := range repos {
			if err := s.apply(ctx, slices.Contains(adopt, s.Name)); err != nil {
				return err
			}
		}
		return nil
	})
}

// apply declares one repository, its dependabot alert switch, and the ruleset
// that keeps its default branch from being force-pushed or deleted. adopt
// switches every resource from create to import, for the first run against a
// repo that already exists. It is set per repo, from the comma-separated
// adopt config, so a run does not re-import repos already in state.
func (s spec) apply(ctx *pulumi.Context, adopt bool) error {
	visibility := "public"
	if s.Private {
		visibility = "private"
	}

	opts := []pulumi.ResourceOption{pulumi.Protect(true)}
	if adopt {
		opts = append(opts, pulumi.Import(pulumi.ID(s.Name)))
	}

	args := &github.RepositoryArgs{
		Name:        pulumi.String(s.Name),
		Description: pulumi.String(s.Description),
		HomepageUrl: pulumi.String(s.Homepage),
		Topics:      pulumi.ToStringArray(s.Topics),
		Visibility:  pulumi.String(visibility),

		HasIssues:   pulumi.Bool(!s.NoIssues),
		HasWiki:     pulumi.Bool(false),
		HasProjects: pulumi.Bool(false),

		AllowMergeCommit:    pulumi.Bool(true),
		AllowSquashMerge:    pulumi.Bool(true),
		AllowRebaseMerge:    pulumi.Bool(false),
		AllowAutoMerge:      pulumi.Bool(true),
		AllowUpdateBranch:   pulumi.Bool(true),
		DeleteBranchOnMerge: pulumi.Bool(true),

		ArchiveOnDestroy: pulumi.Bool(true),
	}

	repo, err := github.NewRepository(ctx, s.Name, args, opts...)
	if err != nil {
		return err
	}

	alertOpts := []pulumi.ResourceOption{pulumi.Parent(repo)}
	if adopt {
		alertOpts = append(alertOpts, pulumi.Import(pulumi.ID(s.Name)))
	}

	_, err = github.NewRepositoryVulnerabilityAlerts(ctx, s.Name+"-alerts", &github.RepositoryVulnerabilityAlertsArgs{
		Repository: repo.Name,
		Enabled:    pulumi.Bool(true),
	}, alertOpts...)
	if err != nil {
		return err
	}

	_, err = github.NewWorkflowRepositoryPermissions(ctx, s.Name+"-workflow-permissions", &github.WorkflowRepositoryPermissionsArgs{
		Repository:                   repo.Name,
		DefaultWorkflowPermissions:   pulumi.String("read"),
		CanApprovePullRequestReviews: pulumi.Bool(s.CanApprovePRs),
	}, pulumi.Parent(repo))
	if err != nil {
		return err
	}

	// Private repos report this as unavailable: automated security fixes need
	// Advanced Security, which is not on a personal plan.
	if !s.Private {
		_, err = github.NewRepositoryDependabotSecurityUpdates(ctx, s.Name+"-dependabot-security-updates", &github.RepositoryDependabotSecurityUpdatesArgs{
			Repository: repo.Name,
			Enabled:    pulumi.Bool(!s.NoDependabotSecurityUpdates),
		}, pulumi.Parent(repo))
		if err != nil {
			return err
		}
	}

	rules := &github.RepositoryRulesetRulesArgs{
		Deletion:       pulumi.Bool(true),
		NonFastForward: pulumi.Bool(true),
	}
	if s.RequirePR {
		rules.PullRequest = &github.RepositoryRulesetRulesPullRequestArgs{
			RequiredApprovingReviewCount: pulumi.Int(0),
			AllowedMergeMethods:          pulumi.StringArray{pulumi.String("squash")},
		}
	}

	rulesetOpts := []pulumi.ResourceOption{pulumi.Parent(repo)}
	if adopt && s.RulesetID != "" {
		rulesetOpts = append(rulesetOpts, pulumi.Import(pulumi.ID(s.Name+":"+s.RulesetID)))
	}

	_, err = github.NewRepositoryRuleset(ctx, s.Name+"-default-branch", &github.RepositoryRulesetArgs{
		Name:        pulumi.String("main guard"),
		Repository:  repo.Name,
		Target:      pulumi.String("branch"),
		Enforcement: pulumi.String("active"),
		Conditions: &github.RepositoryRulesetConditionsArgs{
			RefName: &github.RepositoryRulesetConditionsRefNameArgs{
				Includes: pulumi.StringArray{pulumi.String("~DEFAULT_BRANCH")},
				Excludes: pulumi.StringArray{},
			},
		},
		Rules: rules,
	}, rulesetOpts...)
	return err
}
