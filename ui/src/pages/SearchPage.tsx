import { useState, useCallback } from 'react';
import { Search, Star, Download, Globe, Database, Settings } from 'lucide-react';
import Card from '../components/Card';
import Badge from '../components/Badge';
import HandButton from '../components/HandButton';
import { HandInput, HandSelect } from '../components/HandInput';
import SkillPickerModal from '../components/SkillPickerModal';
import HubManagerModal, { type SavedHub } from '../components/HubManagerModal';
import EmptyState from '../components/EmptyState';
import { useToast } from '../components/Toast';
import { api, type SearchResult, type DiscoveredSkill } from '../api/client';

type SearchMode = 'github' | 'hub';

const LS_MODE = 'search:mode';
const LS_SELECTED = 'search:selectedHub';
const LS_SAVED = 'search:savedHubs';

function loadMode(): SearchMode {
  const v = localStorage.getItem(LS_MODE);
  return v === 'hub' ? 'hub' : 'github';
}

function loadSelectedHub(): string {
  return localStorage.getItem(LS_SELECTED) ?? '';
}

function loadSavedHubs(): SavedHub[] {
  try {
    const raw = localStorage.getItem(LS_SAVED);
    if (!raw) return [];
    return JSON.parse(raw) as SavedHub[];
  } catch {
    return [];
  }
}

function normalizeURL(url: string): string {
  return url.trim().replace(/\/+$/, '');
}

export default function SearchPage() {
  const [mode, setMode] = useState<SearchMode>(loadMode);
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[] | null>(null);
  const [searching, setSearching] = useState(false);
  const { toast } = useToast();

  // Hub state
  const [selectedHub, setSelectedHub] = useState(loadSelectedHub);
  const [savedHubs, setSavedHubs] = useState<SavedHub[]>(loadSavedHubs);
  const [showHubManager, setShowHubManager] = useState(false);

  // Install state
  const [installing, setInstalling] = useState<string | null>(null);

  // Discovery flow state
  const [discoveredSkills, setDiscoveredSkills] = useState<DiscoveredSkill[]>([]);
  const [showPicker, setShowPicker] = useState(false);
  const [pendingSource, setPendingSource] = useState('');
  const [batchInstalling, setBatchInstalling] = useState(false);

  const switchMode = useCallback((newMode: SearchMode) => {
    setMode(newMode);
    setResults(null);
    localStorage.setItem(LS_MODE, newMode);
  }, []);

  const handleSearch = async (searchQuery?: string) => {
    const q = searchQuery ?? query;
    if (mode === 'hub' && !selectedHub) {
      toast('Add a hub source first', 'error');
      return;
    }
    setSearching(true);
    try {
      let res: { results: SearchResult[] };
      if (mode === 'hub') {
        res = await api.searchHub(q, selectedHub);
      } else {
        res = await api.search(q);
      }
      setResults(res.results);
      if (res.results.length === 0) {
        toast(q ? 'No results found.' : 'No skills found.', 'info');
      }
    } catch (e: unknown) {
      toast((e as Error).message, 'error');
    } finally {
      setSearching(false);
    }
  };

  const handleInstall = async (source: string) => {
    setInstalling(source);
    try {
      const disc = await api.discover(source);
      if (disc.skills.length > 1) {
        setDiscoveredSkills(disc.skills);
        setPendingSource(source);
        setShowPicker(true);
      } else if (disc.skills.length === 1) {
        const res = await api.installBatch({ source, skills: disc.skills });
        let hasAuditBlock = false;
        for (const item of res.results) {
          if (item.error) {
            if (item.error.includes('security audit failed')) {
              hasAuditBlock = true;
              toast(`${item.name}: blocked by security audit`, 'error');
            } else {
              toast(`${item.name}: ${item.error}`, 'error');
            }
          }
          if (item.warnings?.length) {
            item.warnings.forEach((w) => toast(`${item.name}: ${w}`, 'warning'));
          }
        }
        toast(res.summary, hasAuditBlock ? 'warning' : 'success');
      } else {
        const res = await api.install({ source });
        toast(
          `Installed: ${res.skillName ?? res.repoName} (${res.action})`,
          'success',
        );
        if (res.warnings?.length > 0) {
          res.warnings.forEach((w) => toast(w, 'warning'));
        }
      }
    } catch (e: unknown) {
      toast((e as Error).message, 'error');
    } finally {
      setInstalling(null);
    }
  };

  const handleBatchInstall = async (selected: DiscoveredSkill[]) => {
    setBatchInstalling(true);
    try {
      const res = await api.installBatch({
        source: pendingSource,
        skills: selected,
      });
      let hasAuditBlock = false;
      for (const item of res.results) {
        if (item.error) {
          if (item.error.includes('security audit failed')) {
            hasAuditBlock = true;
            toast(`${item.name}: blocked by security audit — use Force to override`, 'error');
          } else {
            toast(`${item.name}: ${item.error}`, 'error');
          }
        }
        if (item.warnings?.length) {
          item.warnings.forEach((w) => toast(`${item.name}: ${w}`, 'warning'));
        }
      }
      toast(res.summary, hasAuditBlock ? 'warning' : 'success');
      setShowPicker(false);
    } catch (e: unknown) {
      toast((e as Error).message, 'error');
    } finally {
      setBatchInstalling(false);
    }
  };

  const handleHubsSave = (updated: SavedHub[]) => {
    setSavedHubs(updated);
    localStorage.setItem(LS_SAVED, JSON.stringify(updated));
    // If selected hub was removed, select first available or clear
    if (updated.length === 0) {
      setSelectedHub('');
      localStorage.setItem(LS_SELECTED, '');
    } else if (!updated.some((h) => normalizeURL(h.url) === normalizeURL(selectedHub))) {
      setSelectedHub(updated[0].url);
      localStorage.setItem(LS_SELECTED, updated[0].url);
    }
  };

  const handleSelectHub = (url: string) => {
    setSelectedHub(url);
    localStorage.setItem(LS_SELECTED, url);
    setResults(null);
  };

  return (
    <div className="animate-sketch-in">
      {/* Header */}
      <div className="mb-6">
        <h2
          className="text-3xl md:text-4xl font-bold text-pencil mb-2"
          style={{ fontFamily: 'var(--font-heading)' }}
        >
          Search Skills
        </h2>
        <p className="text-pencil-light">
          Discover and install skills
        </p>
      </div>

      {/* Mode tabs */}
      <div className="flex gap-2 mb-4">
        <HandButton
          onClick={() => switchMode('github')}
          variant={mode === 'github' ? 'primary' : 'secondary'}
          size="sm"
        >
          <Globe size={14} strokeWidth={2.5} />
          GitHub
        </HandButton>
        <HandButton
          onClick={() => switchMode('hub')}
          variant={mode === 'hub' ? 'primary' : 'secondary'}
          size="sm"
        >
          <Database size={14} strokeWidth={2.5} />
          Hub
        </HandButton>
      </div>

      {/* Hub selector (only in hub mode) */}
      {mode === 'hub' && (
        <Card className="mb-4 !overflow-visible">
          {savedHubs.length > 0 ? (
            <div className="flex items-center gap-2">
              <HandSelect
                value={selectedHub}
                onChange={handleSelectHub}
                options={savedHubs.map((h) => ({ value: h.url, label: h.label }))}
                className="flex-1"
              />
              <HandButton
                onClick={() => setShowHubManager(true)}
                variant="ghost"
                size="sm"
                title="Manage hubs"
              >
                <Settings size={14} strokeWidth={2.5} />
                Manage
              </HandButton>
            </div>
          ) : (
            <div className="text-center py-3">
              <p className="text-base text-muted-dark mb-3">
                No hubs configured. Add one to get started.
              </p>
              <HandButton
                onClick={() => setShowHubManager(true)}
                variant="secondary"
                size="sm"
              >
                <Settings size={14} strokeWidth={2.5} />
                Manage Hubs
              </HandButton>
            </div>
          )}
        </Card>
      )}

      {/* Search box */}
      <Card className="mb-6">
        <div className="flex gap-3">
          <div className="relative flex-1">
            <Search
              size={18}
              strokeWidth={2.5}
              className="absolute left-4 top-1/2 -translate-y-1/2 text-muted-dark pointer-events-none"
            />
            <HandInput
              type="text"
              placeholder={mode === 'github' ? 'Search GitHub for skills...' : 'Search hub...'}
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch(query)}
              className="!pl-11"
            />
          </div>
          <HandButton
            onClick={() => handleSearch(query)}
            disabled={searching}
            variant="primary"
            size="md"
          >
            <Search size={16} strokeWidth={2.5} />
            {searching ? 'Searching...' : 'Search'}
          </HandButton>
        </div>
        {mode === 'github' && (
          <p className="text-sm text-muted-dark mt-3 flex items-center gap-1">
            <Globe size={12} strokeWidth={2} />
            Requires GITHUB_TOKEN environment variable for GitHub API access.
          </p>
        )}
      </Card>

      {/* Results */}
      {results && results.length > 0 && (
        <div className="space-y-4">
          <p className="text-base text-pencil-light">
            {results.length} result{results.length !== 1 ? 's' : ''} found
          </p>
          {results.map((r, i) => (
            <Card
              key={r.source}
              className={i % 2 === 0 ? 'rotate-[-0.15deg]' : 'rotate-[0.15deg]'}
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1 flex-wrap">
                    <span
                      className="font-bold text-pencil"
                      style={{ fontFamily: 'var(--font-heading)' }}
                    >
                      {r.name}
                    </span>
                    {r.stars > 0 && (
                      <span className="flex items-center gap-1 text-sm text-warning">
                        <Star size={14} strokeWidth={2.5} fill="currentColor" />
                        {r.stars}
                      </span>
                    )}
                    {r.owner && <Badge>{r.owner}</Badge>}
                  </div>
                  {r.description && (
                    <p className="text-base text-pencil-light mb-1.5">{r.description}</p>
                  )}
                  <p
                    className="text-sm text-muted-dark truncate"
                    style={{ fontFamily: "'Courier New', monospace" }}
                  >
                    {r.source}
                  </p>
                </div>
                <HandButton
                  onClick={() => handleInstall(r.source)}
                  disabled={installing === r.source}
                  variant="secondary"
                  size="sm"
                  className="shrink-0"
                >
                  <Download size={14} strokeWidth={2.5} />
                  {installing === r.source ? 'Installing...' : 'Install'}
                </HandButton>
              </div>
            </Card>
          ))}
        </div>
      )}

      {results && results.length === 0 && (
        <EmptyState
          icon={Search}
          title="No results found"
          description={
            mode === 'github'
              ? 'Try different search terms or check your GITHUB_TOKEN.'
              : 'Try different search terms or check your hub source.'
          }
        />
      )}

      {/* Initial state before any search */}
      {!results && !searching && (
        <div className="text-center py-12">
          <div
            className="inline-flex items-center justify-center w-20 h-20 bg-postit border-2 border-dashed border-pencil-light mb-4"
            style={{ borderRadius: '50%' }}
          >
            <Search size={32} strokeWidth={2} className="text-pencil-light" />
          </div>
          <p
            className="text-xl text-pencil mb-1"
            style={{ fontFamily: 'var(--font-heading)' }}
          >
            Start searching
          </p>
          <p className="text-base text-pencil-light mb-4">
            {mode === 'github'
              ? 'Type a query above to find skills on GitHub'
              : 'Type a query above, or search with empty query to browse all'}
          </p>
          <HandButton
            onClick={() => handleSearch('')}
            variant="secondary"
            size="sm"
          >
            <Star size={14} strokeWidth={2.5} />
            {mode === 'github' ? 'Browse Popular Skills' : 'Browse All Skills'}
          </HandButton>
        </div>
      )}

      {/* Hub Manager Modal */}
      <HubManagerModal
        open={showHubManager}
        hubs={savedHubs}
        onSave={handleHubsSave}
        onClose={() => setShowHubManager(false)}
      />

      {/* Skill Picker Modal for multi-skill repos */}
      <SkillPickerModal
        open={showPicker}
        source={pendingSource}
        skills={discoveredSkills}
        onInstall={handleBatchInstall}
        onCancel={() => setShowPicker(false)}
        installing={batchInstalling}
      />
    </div>
  );
}
