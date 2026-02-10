import { useState, useEffect, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { Save, FileCode, ShieldCheck, ArrowLeft, FilePlus } from 'lucide-react';
import CodeMirror from '@uiw/react-codemirror';
import { yaml } from '@codemirror/lang-yaml';
import { EditorView } from '@codemirror/view';
import Card from '../components/Card';
import HandButton from '../components/HandButton';
import EmptyState from '../components/EmptyState';
import { PageSkeleton } from '../components/Skeleton';
import { useToast } from '../components/Toast';
import { api } from '../api/client';
import { useApi } from '../hooks/useApi';
import { useAppContext } from '../context/AppContext';
import { handTheme } from '../lib/codemirror-theme';

export default function AuditRulesPage() {
  const { data, loading, error, refetch } = useApi(() => api.getAuditRules());
  const [raw, setRaw] = useState('');
  const [saving, setSaving] = useState(false);
  const [creating, setCreating] = useState(false);
  const [dirty, setDirty] = useState(false);
  const { toast } = useToast();
  const { isProjectMode } = useAppContext();

  const extensions = useMemo(() => [yaml(), EditorView.lineWrapping], []);

  useEffect(() => {
    if (data?.raw) {
      setRaw(data.raw);
      setDirty(false);
    }
  }, [data]);

  const handleChange = (value: string) => {
    setRaw(value);
    setDirty(value !== (data?.raw ?? ''));
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await api.putAuditRules(raw);
      toast('Audit rules saved successfully.', 'success');
      setDirty(false);
      refetch();
    } catch (e: unknown) {
      toast((e as Error).message, 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleCreate = async () => {
    setCreating(true);
    try {
      await api.initAuditRules();
      toast('Audit rules file created.', 'success');
      refetch();
    } catch (e: unknown) {
      toast((e as Error).message, 'error');
    } finally {
      setCreating(false);
    }
  };

  if (loading) return <PageSkeleton />;
  if (error) {
    return (
      <Card variant="accent" className="text-center py-8">
        <p className="text-danger text-lg" style={{ fontFamily: 'var(--font-heading)' }}>
          Failed to load audit rules
        </p>
        <p className="text-pencil-light text-sm mt-1">{error}</p>
      </Card>
    );
  }

  return (
    <div className="animate-sketch-in">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Link
              to="/audit"
              className="text-pencil-light hover:text-pencil transition-colors"
            >
              <ArrowLeft size={20} strokeWidth={2.5} />
            </Link>
            <h2
              className="text-3xl font-bold text-pencil flex items-center gap-2"
              style={{ fontFamily: 'var(--font-heading)' }}
            >
              <ShieldCheck size={28} strokeWidth={2.5} />
              Custom Audit Rules
            </h2>
          </div>
          <p
            className="text-pencil-light"
            style={{ fontFamily: 'var(--font-hand)' }}
          >
            {isProjectMode
              ? 'Project-level custom rules for security scanning'
              : 'Global custom rules for security scanning'}
          </p>
        </div>
        {data?.exists && (
          <div className="flex items-center gap-3">
            {dirty && (
              <span
                className="text-sm text-warning px-2 py-1 bg-warning-light rounded-full border border-warning"
                style={{ fontFamily: 'var(--font-hand)' }}
              >
                unsaved changes
              </span>
            )}
            <HandButton
              onClick={handleSave}
              disabled={saving || !dirty}
              variant="primary"
              size="md"
            >
              <Save size={16} strokeWidth={2.5} />
              {saving ? 'Saving...' : 'Save'}
            </HandButton>
          </div>
        )}
      </div>

      {/* File does not exist */}
      {data && !data.exists && (
        <EmptyState
          icon={FilePlus}
          title="No custom rules file"
          description={`Create ${isProjectMode ? 'a project-level' : 'a global'} audit-rules.yaml to add or override security rules`}
          action={
            <HandButton variant="primary" onClick={handleCreate} disabled={creating}>
              <FilePlus size={16} strokeWidth={2.5} />
              {creating ? 'Creating...' : 'Create Rules File'}
            </HandButton>
          }
        />
      )}

      {/* File exists — editor */}
      {data?.exists && (
        <Card decoration="tape">
          <div className="flex items-center gap-2 mb-3">
            <FileCode size={16} strokeWidth={2.5} className="text-blue" />
            <span
              className="text-base text-pencil-light"
              style={{ fontFamily: 'var(--font-hand)' }}
            >
              {data.path}
            </span>
          </div>
          <div className="min-w-0 -mx-4 -mb-4">
            <CodeMirror
              value={raw}
              onChange={handleChange}
              extensions={extensions}
              theme={handTheme}
              height="500px"
              basicSetup={{
                lineNumbers: true,
                foldGutter: true,
                highlightActiveLine: true,
                highlightSelectionMatches: true,
                bracketMatching: true,
                indentOnInput: true,
                autocompletion: false,
              }}
            />
          </div>
        </Card>
      )}
    </div>
  );
}
