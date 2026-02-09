import { useState, useCallback } from 'react';
import { Download, Package, ChevronDown, ChevronUp, ShieldAlert, ShieldCheck } from 'lucide-react';
import Card from './Card';
import HandButton from './HandButton';
import Badge from './Badge';
import { HandInput, HandCheckbox } from './HandInput';
import SkillPickerModal from './SkillPickerModal';
import ConfirmDialog from './ConfirmDialog';
import { useToast } from './Toast';
import { api, type InstallResult, type DiscoveredSkill } from '../api/client';

interface InstallFormProps {
  /** Called after a successful install with the result */
  onSuccess?: (result: InstallResult) => void;
  /** Whether the form starts expanded (default: false) */
  defaultOpen?: boolean;
  /** Whether to show the collapsible toggle header (default: true) */
  collapsible?: boolean;
  className?: string;
}

/** Parse audit error message into individual finding lines */
function parseAuditError(msg: string): string[] {
  return msg
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l.startsWith('CRITICAL:') || l.startsWith('HIGH:') || l.startsWith('"'));
}

/** Check if an error is an audit block */
function isAuditBlock(msg: string): boolean {
  return msg.includes('security audit failed');
}

/** Saved install params for force-retry */
interface PendingInstall {
  type: 'single' | 'batch' | 'track';
  source: string;
  name?: string;
  skills?: DiscoveredSkill[];
}

export default function InstallForm({
  onSuccess,
  defaultOpen = false,
  collapsible = true,
  className = '',
}: InstallFormProps) {
  const [open, setOpen] = useState(defaultOpen);
  const [source, setSource] = useState('');
  const [name, setName] = useState('');
  const [track, setTrack] = useState(false);
  const [force, setForce] = useState(false);
  const [installing, setInstalling] = useState(false);
  const { toast } = useToast();

  // Discovery flow state
  const [discoveredSkills, setDiscoveredSkills] = useState<DiscoveredSkill[]>([]);
  const [showPicker, setShowPicker] = useState(false);
  const [pendingSource, setPendingSource] = useState('');
  const [batchInstalling, setBatchInstalling] = useState(false);

  // Audit dialog state
  const [auditDialog, setAuditDialog] = useState<{
    findings: string[];
    pending: PendingInstall;
  } | null>(null);
  const [auditForcing, setAuditForcing] = useState(false);
  const [warningDialog, setWarningDialog] = useState<string[] | null>(null);

  const resetForm = () => {
    setSource('');
    setName('');
    setTrack(false);
    setForce(false);
    if (collapsible) setOpen(false);
  };

  /** Handle install result: show warning dialog if warnings exist, otherwise just toast */
  const handleResult = useCallback(
    (res: InstallResult, label?: string) => {
      const prefix = label ? `${label}: ` : '';
      toast(`${prefix}Installed (${res.action})`, 'success');
      if (res.warnings && res.warnings.length > 0) {
        setWarningDialog(res.warnings);
      }
      resetForm();
      onSuccess?.(res);
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [onSuccess, toast],
  );

  /** Handle error: if audit block, show confirm dialog; otherwise error toast */
  const handleError = useCallback(
    (e: unknown, pending: PendingInstall) => {
      const msg = (e as Error).message;
      if (isAuditBlock(msg)) {
        setAuditDialog({ findings: parseAuditError(msg), pending });
      } else {
        toast(msg, 'error');
      }
    },
    [toast],
  );

  /** Force-install after audit confirm */
  const handleAuditForce = async () => {
    if (!auditDialog) return;
    const { pending } = auditDialog;
    setAuditForcing(true);
    try {
      if (pending.type === 'track') {
        const res = await api.install({
          source: pending.source,
          name: pending.name,
          track: true,
          force: true,
        });
        handleResult(res, res.skillName ?? res.repoName);
      } else if (pending.type === 'batch') {
        const res = await api.installBatch({
          source: pending.source,
          skills: pending.skills!,
          force: true,
        });
        toast(res.summary, 'success');
        const allWarnings: string[] = [];
        for (const item of res.results) {
          if (item.error) toast(`${item.name}: ${item.error}`, 'error');
          if (item.warnings?.length) allWarnings.push(...item.warnings.map((w) => `${item.name}: ${w}`));
        }
        if (allWarnings.length > 0) setWarningDialog(allWarnings);
        resetForm();
        onSuccess?.({ action: 'installed', warnings: [], skillName: res.summary });
      } else {
        const res = await api.install({
          source: pending.source,
          name: pending.name,
          force: true,
        });
        handleResult(res, res.skillName ?? res.repoName);
      }
    } catch (e: unknown) {
      toast((e as Error).message, 'error');
    } finally {
      setAuditForcing(false);
      setAuditDialog(null);
    }
  };

  const handleInstall = async () => {
    if (!source.trim()) return;
    const trimmed = source.trim();

    // Track mode → direct install (no discovery needed)
    if (track) {
      setInstalling(true);
      try {
        const res = await api.install({
          source: trimmed,
          name: name.trim() || undefined,
          track: true,
          force,
        });
        handleResult(res, res.skillName ?? res.repoName);
      } catch (e: unknown) {
        handleError(e, { type: 'track', source: trimmed, name: name.trim() || undefined });
      } finally {
        setInstalling(false);
      }
      return;
    }

    // Discovery flow
    setInstalling(true);
    try {
      const disc = await api.discover(trimmed);
      if (disc.skills.length > 1) {
        // Multiple skills found — open picker
        setDiscoveredSkills(disc.skills);
        setPendingSource(trimmed);
        setShowPicker(true);
      } else if (disc.skills.length === 1) {
        // Single discovered skill — install via batch
        const res = await api.installBatch({
          source: trimmed,
          skills: disc.skills,
          force,
        });
        const allWarnings: string[] = [];
        let hasError = false;
        for (const item of res.results) {
          if (item.error) {
            hasError = true;
            // Check if any batch item was audit-blocked
            if (isAuditBlock(item.error)) {
              setAuditDialog({
                findings: parseAuditError(item.error),
                pending: { type: 'batch', source: trimmed, skills: disc.skills },
              });
              return;
            }
            toast(`${item.name}: ${item.error}`, 'error');
          }
          if (item.warnings?.length) allWarnings.push(...item.warnings.map((w) => `${item.name}: ${w}`));
        }
        if (!hasError) toast(res.summary, 'success');
        if (allWarnings.length > 0) setWarningDialog(allWarnings);
        resetForm();
        onSuccess?.({ action: 'installed', warnings: [], skillName: res.summary });
      } else {
        // No skills discovered — direct install
        const res = await api.install({
          source: trimmed,
          name: name.trim() || undefined,
          force,
        });
        handleResult(res, res.skillName ?? res.repoName);
      }
    } catch (e: unknown) {
      handleError(e, { type: 'single', source: trimmed, name: name.trim() || undefined });
    } finally {
      setInstalling(false);
    }
  };

  const handleBatchInstall = async (selected: DiscoveredSkill[]) => {
    setBatchInstalling(true);
    try {
      const res = await api.installBatch({
        source: pendingSource,
        skills: selected,
        force,
      });
      const allWarnings: string[] = [];
      let hasError = false;
      for (const item of res.results) {
        if (item.error) {
          hasError = true;
          if (isAuditBlock(item.error)) {
            setShowPicker(false);
            setAuditDialog({
              findings: parseAuditError(item.error),
              pending: { type: 'batch', source: pendingSource, skills: selected },
            });
            return;
          }
          toast(`${item.name}: ${item.error}`, 'error');
        }
        if (item.warnings?.length) allWarnings.push(...item.warnings.map((w) => `${item.name}: ${w}`));
      }
      if (!hasError) toast(res.summary, 'success');
      if (allWarnings.length > 0) setWarningDialog(allWarnings);
      setShowPicker(false);
      resetForm();
      onSuccess?.({ action: 'installed', warnings: [], skillName: res.summary });
    } catch (e: unknown) {
      handleError(e, { type: 'batch', source: pendingSource, skills: selected });
    } finally {
      setBatchInstalling(false);
    }
  };

  const formContent = (
    <Card variant="postit" className="animate-sketch-in">
      <div className="space-y-4">
        <HandInput
          label="Source (GitHub URL, owner/repo, or local path)"
          type="text"
          placeholder="owner/repo or https://github.com/..."
          value={source}
          onChange={(e) => setSource(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleInstall()}
        />
        <HandInput
          label="Name override (optional)"
          type="text"
          placeholder="custom-name"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <div className="flex items-center gap-6">
          <HandCheckbox
            label="Track (git repo)"
            checked={track}
            onChange={setTrack}
          />
          <HandCheckbox
            label="Force overwrite"
            checked={force}
            onChange={setForce}
          />
        </div>
        <HandButton
          onClick={handleInstall}
          disabled={installing || !source.trim()}
          variant="primary"
          size="sm"
        >
          <Download size={14} strokeWidth={2.5} />
          {installing ? 'Installing...' : 'Install'}
        </HandButton>
      </div>
    </Card>
  );

  const pickerModal = (
    <SkillPickerModal
      open={showPicker}
      source={pendingSource}
      skills={discoveredSkills}
      onInstall={handleBatchInstall}
      onCancel={() => setShowPicker(false)}
      installing={batchInstalling}
    />
  );

  const auditConfirmDialog = (
    <ConfirmDialog
      open={!!auditDialog}
      variant="danger"
      wide
      title="Security Threats Detected"
      message={
        <div className="text-left space-y-2">
          <div className="flex items-center gap-2 justify-center mb-3">
            <ShieldAlert size={20} className="text-danger" />
            <span>Critical issues found during security audit</span>
          </div>
          <div
            className="bg-paper border border-danger/30 p-3 space-y-1 text-sm text-pencil font-mono max-h-48 overflow-y-auto"
            style={{ borderRadius: '6px' }}
          >
            {auditDialog?.findings.map((line, i) => (
              <div key={i} className={line.startsWith('"') ? 'text-pencil-light pl-4' : ''}>
                {line.startsWith('CRITICAL:') ? (
                  <span><Badge variant="danger">CRITICAL</Badge> {line.replace('CRITICAL: ', '')}</span>
                ) : line.startsWith('HIGH:') ? (
                  <span><Badge variant="warning">HIGH</Badge> {line.replace('HIGH: ', '')}</span>
                ) : (
                  line
                )}
              </div>
            ))}
          </div>
          <p className="text-xs text-pencil-light mt-2">
            Force install will bypass the security check. Proceed with caution.
          </p>
        </div>
      }
      confirmText="Force Install"
      cancelText="Cancel"
      onConfirm={handleAuditForce}
      onCancel={() => setAuditDialog(null)}
      loading={auditForcing}
    />
  );

  const warningConfirmDialog = (
    <ConfirmDialog
      open={!!warningDialog}
      variant="default"
      wide
      title="Security Warnings"
      message={
        <div className="text-left space-y-2">
          <div className="flex items-center gap-2 justify-center mb-3">
            <ShieldCheck size={20} className="text-warning" />
            <span>Skill installed with audit warnings</span>
          </div>
          <div
            className="bg-paper border border-warning/30 p-3 space-y-2 text-sm text-pencil font-mono max-h-48 overflow-y-auto"
            style={{ borderRadius: '6px' }}
          >
            {warningDialog?.map((w, i) => {
              const lines = w.split('\n');
              const header = lines[0];
              const snippet = lines.slice(1).map((l) => l.trim()).filter(Boolean).join(' ');
              const isHigh = header.includes('HIGH');
              return (
                <div key={i}>
                  <div>
                    <Badge variant={isHigh ? 'warning' : 'info'}>
                      {isHigh ? 'HIGH' : 'MEDIUM'}
                    </Badge>{' '}
                    {header.replace(/^audit\s+(HIGH|MEDIUM|CRITICAL):\s*/, '')}
                  </div>
                  {snippet && <div className="text-pencil-light pl-4 text-xs">{snippet}</div>}
                </div>
              );
            })}
          </div>
        </div>
      }
      confirmText="OK"
      cancelText=""
      onConfirm={() => setWarningDialog(null)}
      onCancel={() => setWarningDialog(null)}
    />
  );

  if (!collapsible) {
    return (
      <div className={className}>
        {formContent}
        {pickerModal}
        {auditConfirmDialog}
        {warningConfirmDialog}
      </div>
    );
  }

  return (
    <div className={className}>
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-2 text-pencil-light hover:text-pencil transition-colors cursor-pointer mb-3"
        style={{
          fontFamily: 'var(--font-hand)',
          background: 'none',
          border: 'none',
          padding: 0,
        }}
      >
        <Package size={16} strokeWidth={2.5} />
        <span className="text-base">Install from URL / Path</span>
        {open ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
      </button>
      {open && formContent}
      {pickerModal}
      {auditConfirmDialog}
      {warningConfirmDialog}
    </div>
  );
}
