import { useState, useEffect } from 'react';
import './App.css';
import { GetCopilotUsage } from '../wailsjs/go/main/App';

interface UsageReport {
  provider: string;
  periodStart?: string;
  periodEnd?: string;
  retrievedAt: string;
  metrics: Record<string, number>;
  metadata?: Record<string, string>;
}

function pct(value: number | undefined): string {
  if (value === undefined) return '—';
  return `${Math.round(value)}%`;
}

function MetricRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="metric-row">
      <span className="metric-label">{label}</span>
      <span className="metric-value">{value}</span>
    </div>
  );
}

function App() {
  const [report, setReport] = useState<UsageReport | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  async function refresh() {
    setLoading(true);
    setError(null);
    try {
      const r = await GetCopilotUsage();
      setReport(r);
    } catch (e: unknown) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { refresh(); }, []);

  const m = report?.metrics ?? {};
  const meta = report?.metadata ?? {};

  return (
    <div id="App">
      <header className="app-header">
        <h1>ingo</h1>
        <p className="app-subtitle">AI token usage monitor</p>
      </header>

      <main className="usage-card">
        {loading && <p className="status-text">Loading…</p>}

        {!loading && error && (
          <div className="error-box">
            <p className="error-title">Could not fetch Copilot usage</p>
            <p className="error-detail">{error}</p>
          </div>
        )}

        {!loading && report && !error && (
          <>
            <div className="plan-row">
              <span className="plan-badge">{meta.plan ?? 'GitHub Copilot'}</span>
              {meta.quota_reset_date && (
                <span className="reset-date">Resets {meta.quota_reset_date}</span>
              )}
            </div>

            <div className="metrics">
              <MetricRow
                label="Premium interactions remaining"
                value={
                  m['premium_unlimited'] === 1
                    ? 'Unlimited'
                    : pct(m['premium_percent_remaining'])
                }
              />
              <MetricRow
                label="Chat remaining"
                value={
                  m['chat_unlimited'] === 1
                    ? 'Unlimited'
                    : pct(m['chat_percent_remaining'])
                }
              />
              <MetricRow
                label="Completions remaining"
                value={
                  m['completions_unlimited'] === 1
                    ? 'Unlimited'
                    : pct(m['completions_percent_remaining'])
                }
              />
            </div>

            <p className="retrieved-at">
              Updated {new Date(report.retrievedAt).toLocaleTimeString()}
            </p>
          </>
        )}
      </main>

      <button className="refresh-btn" onClick={refresh} disabled={loading}>
        {loading ? 'Refreshing…' : 'Refresh'}
      </button>
    </div>
  );
}

export default App;
