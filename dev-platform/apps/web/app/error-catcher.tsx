'use client';

import { useEffect, useState } from 'react';

export default function ErrorCatcher() {
  const [errs, setErrs] = useState<string[]>([]);

  useEffect(() => {
    const onError = (e: ErrorEvent) => {
      const msg = `[window error] ${e.message || 'unknown'} @ ${e.filename || ''}:${e.lineno || ''}:${e.colno || ''}`;
      setErrs((prev) => [...prev.slice(-4), msg]);
    };
    const onReject = (e: PromiseRejectionEvent) => {
      const r = e.reason as { message?: string; stack?: string } | string | null | undefined;
      const msg = `[unhandledrejection] ${(r && typeof r === 'object' && r.message) || r || ''}`;
      setErrs((prev) => [...prev.slice(-4), msg]);
    };
    window.addEventListener('error', onError);
    window.addEventListener('unhandledrejection', onReject);
    return () => {
      window.removeEventListener('error', onError);
      window.removeEventListener('unhandledrejection', onReject);
    };
  }, []);

  if (errs.length === 0) return null;
  return (
    <div
      style={{
        position: 'fixed',
        bottom: 0,
        left: 0,
        right: 0,
        zIndex: 99999,
        background: '#111',
        color: '#f87171',
        fontFamily: 'ui-monospace, monospace',
        fontSize: 11,
        padding: 8,
        maxHeight: 200,
        overflow: 'auto',
      }}
    >
      <div style={{ color: '#fbbf24', fontWeight: 700, marginBottom: 4 }}>Client errors detected:</div>
      {errs.map((e, i) => (
        <div key={i}>{e}</div>
      ))}
    </div>
  );
}
