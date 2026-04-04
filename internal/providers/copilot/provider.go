package copilot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"ingo/internal/usage"
)

const (
	ProviderName   = "github-copilot"
	defaultBaseURL = "https://api.github.com"
)

// TokenResolver provides an access token at fetch time.
type TokenResolver func(ctx context.Context) (string, error)

// Config wires dependencies for the provider.
type Config struct {
	HTTPClient    *http.Client
	BaseURL       string
	TokenResolver TokenResolver
}

// Provider implements usage.Provider for the GitHub Copilot internal user endpoint.
type Provider struct {
	httpClient    *http.Client
	baseURL       string
	tokenResolver TokenResolver
}

// quotaSnapshot represents the per-feature quota data returned by the API.
type quotaSnapshot struct {
	PercentRemaining *float64 `json:"percent_remaining"`
	Remaining        *float64 `json:"remaining"`
	QuotaRemaining   *float64 `json:"quota_remaining"`
	Unlimited        *bool    `json:"unlimited"`
	TimestampUTC     string   `json:"timestamp_utc"`
}

// quotaSnapshots groups all feature snapshots.
type quotaSnapshots struct {
	Chat                *quotaSnapshot `json:"chat"`
	Completions         *quotaSnapshot `json:"completions"`
	PremiumInteractions *quotaSnapshot `json:"premium_interactions"`
}

// apiResponse models the /copilot_internal/user response shape.
type apiResponse struct {
	CopilotPlan      string          `json:"copilot_plan"`
	QuotaResetDate   string          `json:"quota_reset_date"`
	QuotaSnapshots   *quotaSnapshots `json:"quota_snapshots"`
}

// NewProvider constructs a Provider with the given Config.
func NewProvider(cfg Config) *Provider {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	baseURL := cfg.BaseURL
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}

	resolver := cfg.TokenResolver
	if resolver == nil {
		resolver = EnvTokenResolver
	}

	return &Provider{
		httpClient:    client,
		baseURL:       strings.TrimRight(baseURL, "/"),
		tokenResolver: resolver,
	}
}

// Name returns the canonical provider identifier.
func (p *Provider) Name() string {
	return ProviderName
}

// FetchUsage queries /copilot_internal/user and normalises the response.
func (p *Provider) FetchUsage(ctx context.Context) (*usage.UsageReport, error) {
	token, err := p.tokenResolver(ctx)
	if err != nil {
		return nil, err
	}

	endpoint := p.baseURL + "/copilot_internal/user"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Editor-Version", "vscode/1.96.2")
	req.Header.Set("User-Agent", "GitHubCopilotChat/0.26.7")
	req.Header.Set("X-Github-Api-Version", "2025-04-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch Copilot usage: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read Copilot usage response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		hint := ""
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			hint = " (check that GITHUB_COPILOT_TOKEN is valid)"
		}
		return nil, fmt.Errorf("Copilot API returned %s%s: %s",
			resp.Status, hint, strings.TrimSpace(string(body)))
	}

	var payload apiResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode Copilot response: %w", err)
	}

	report := &usage.UsageReport{
		Provider:    p.Name(),
		RetrievedAt: time.Now().UTC().Format(time.RFC3339),
		Metrics:     make(map[string]float64),
		Metadata: map[string]string{
			"endpoint": "/copilot_internal/user",
		},
	}

	if payload.CopilotPlan != "" {
		report.Metadata["plan"] = payload.CopilotPlan
	}
	if payload.QuotaResetDate != "" {
		report.Metadata["quota_reset_date"] = payload.QuotaResetDate
	}

	if qs := payload.QuotaSnapshots; qs != nil {
		applySnapshot(report, "premium", qs.PremiumInteractions)
		applySnapshot(report, "chat", qs.Chat)
		applySnapshot(report, "completions", qs.Completions)
	}

	return report, nil
}

// applySnapshot writes a quota snapshot's fields into the report using a
// consistent key prefix, e.g. "premium_percent_remaining".
// If the quota is unlimited, a sentinel metric of 1 is written for
// "<prefix>_unlimited" and the percentage metrics are omitted.
func applySnapshot(report *usage.UsageReport, prefix string, snap *quotaSnapshot) {
	if snap == nil {
		return
	}
	if snap.Unlimited != nil && *snap.Unlimited {
		report.Metrics[prefix+"_unlimited"] = 1
		return
	}
	addMetricF(report.Metrics, prefix+"_percent_remaining", snap.PercentRemaining)
	addMetricF(report.Metrics, prefix+"_remaining", snap.Remaining)
	addMetricF(report.Metrics, prefix+"_quota_remaining", snap.QuotaRemaining)
}

// EnvTokenResolver resolves a GitHub Copilot token from the environment.
// Preference order: GITHUB_COPILOT_TOKEN > GITHUB_TOKEN > GH_TOKEN.
func EnvTokenResolver(_ context.Context) (string, error) {
	for _, envVar := range []string{"GITHUB_COPILOT_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"} {
		token := strings.TrimSpace(os.Getenv(envVar))
		if token != "" {
			return token, nil
		}
	}
	return "", errors.New(
		"missing token: set GITHUB_COPILOT_TOKEN (preferred), GITHUB_TOKEN, or GH_TOKEN",
	)
}

func addMetricF(metrics map[string]float64, key string, value *float64) {
	if value != nil {
		metrics[key] = *value
	}
}
