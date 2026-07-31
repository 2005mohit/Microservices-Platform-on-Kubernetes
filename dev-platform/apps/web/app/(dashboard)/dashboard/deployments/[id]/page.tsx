'use client';

import { useState, useEffect, useRef } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft, Loader2 } from 'lucide-react';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

const STATUS_COLORS: Record<string, string> = {
  queued: 'bg-yellow-500',
  building: 'bg-blue-500',
  deploying: 'bg-violet-500',
  success: 'bg-emerald-500',
  failed: 'bg-red-500',
};

const STATUS_LABELS: Record<string, string> = {
  queued: 'Queued',
  building: 'Building',
  deploying: 'Deploying',
  success: 'Success',
  failed: 'Failed',
};

export default function DeploymentDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const [logs, setLogs] = useState<string[]>([]);
  const [status, setStatus] = useState('queued');
  const [deployment, setDeployment] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const terminalRef = useRef<HTMLDivElement>(null);

  const doneRef = useRef(false);

  useEffect(() => {
    let cancelled = false;
    function load() {
      fetch(`${API_BASE}/api/v1/deployments/${id}`)
        .then((r) => r.json())
        .then((d) => {
          if (cancelled) return;
          setDeployment(d);
          if (d.logs) setLogs(d.logs.split('\n').filter(Boolean));
        })
        .catch(console.error)
        .finally(() => setLoading(false));
    }
    load();
    const interval = setInterval(() => {
      if (!doneRef.current) load();
    }, 3000);
    return () => { cancelled = true; clearInterval(interval); };
  }, [id]);

  useEffect(() => {
    if (status === 'success' || status === 'failed') doneRef.current = true;
  }, [status]);

  useEffect(() => {
    let ws: WebSocket;
    let reconnectTimer: NodeJS.Timeout;

    function connect() {
      const wsProtocol = API_BASE.startsWith('https') ? 'wss:' : 'ws:';
      const wsHost = API_BASE.replace(/^https?:\/\//, '');
      ws = new WebSocket(`${wsProtocol}//${wsHost}/ws/deployments/${id}`);

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.status) {
            setStatus(data.status);
            if (data.status === 'success' || data.status === 'failed') {
              fetch(`${API_BASE}/api/v1/deployments/${id}`)
                .then((r) => r.json())
                .then((d) => setDeployment(d))
                .catch(() => {});
            }
          }
          if (data.line) setLogs((prev) => [...prev, data.line]);
          if (data.log) setLogs((prev) => [...prev, data.log]);
        } catch {
          setLogs((prev) => [...prev, event.data]);
        }
      };

      ws.onclose = () => {
        if (status !== 'success' && status !== 'failed') {
          reconnectTimer = setTimeout(connect, 3000);
        }
      };

      ws.onerror = () => ws.close();
    }

    connect();
    return () => {
      clearTimeout(reconnectTimer);
      ws?.close();
    };
  }, [id, status]);

  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [logs]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div>
        <Link
          href="/dashboard/projects"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors mb-4"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Projects
        </Link>
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Deployment</h1>
            <p className="text-muted-foreground mt-1 font-mono text-sm">{id}</p>
          </div>
          <div className="flex items-center gap-3">
            <span className={`h-2.5 w-2.5 rounded-full ${STATUS_COLORS[status] || 'bg-gray-500'} ${status === 'building' || status === 'deploying' ? 'animate-pulse' : ''}`} />
            <span className="text-sm font-medium">{STATUS_LABELS[status] || status}</span>
          </div>
        </div>
      </div>

      {deployment?.domain && (
        <div className="rounded-xl border border-border bg-card p-4 flex items-center justify-between">
          <div>
            <p className="text-sm text-muted-foreground">Deployed URL</p>
            <a
              href={`http://${deployment.domain}`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary font-medium hover:underline"
            >
              http://{deployment.domain}
            </a>
          </div>
          <span className={`text-xs px-2 py-1 rounded-full ${status === 'success' ? 'bg-emerald-500/10 text-emerald-500' : 'bg-yellow-500/10 text-yellow-500'}`}>
            {status === 'success' ? 'Live' : 'Pending'}
          </span>
        </div>
      )}

      <div className="rounded-xl border border-border overflow-hidden">
        <div className="bg-black/80 px-4 py-2.5 flex items-center gap-2 border-b border-white/5">
          <div className="flex gap-1.5">
            <div className="h-3 w-3 rounded-full bg-red-500/80" />
            <div className="h-3 w-3 rounded-full bg-yellow-500/80" />
            <div className="h-3 w-3 rounded-full bg-emerald-500/80" />
          </div>
          <span className="text-xs text-white/40 font-mono ml-3">build-log.txt</span>
        </div>
        <div
          ref={terminalRef}
          className="bg-black/90 p-4 h-96 overflow-y-auto font-mono text-sm leading-relaxed terminal-scrollbar"
        >
          {logs.length === 0 ? (
            <span className="text-white/30">Waiting for logs...</span>
          ) : (
            logs.map((line, i) => {
              const isError = line.toLowerCase().includes('error') || line.toLowerCase().includes('failed');
              const isSuccess = line.toLowerCase().includes('success');
              const isBuilding = line.toLowerCase().includes('step') || line.toLowerCase().includes('building');
              return (
                <div
                  key={i}
                  className={`${
                    isError ? 'text-red-400' : isSuccess ? 'text-emerald-400' : isBuilding ? 'text-blue-400' : 'text-white/80'
                  }`}
                >
                  <span className="text-white/20 mr-3 select-none">{String(i + 1).padStart(3, '0')}</span>
                  {line}
                </div>
              );
            })
          )}
          {(status === 'building' || status === 'deploying') && (
            <div className="flex items-center gap-2 text-white/40 mt-2">
              <Loader2 className="w-3 h-3 animate-spin" />
              Running...
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
