'use client';

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html lang="en">
      <body>
        <div style={{ padding: 32, fontFamily: 'ui-monospace, SFMono-Regular, monospace', lineHeight: 1.6 }}>
          <h2 style={{ fontSize: 18, fontWeight: 700 }}>Application error</h2>
          <p style={{ color: '#9ca3af' }}>digest: {error.digest || 'n/a'}</p>
          <p style={{ color: '#ef4444' }}>{error.message}</p>
          {error.stack ? (
            <pre style={{ whiteSpace: 'pre-wrap', fontSize: 12, color: '#6b7280', background: '#111', padding: 12, borderRadius: 8 }}>
              {error.stack}
            </pre>
          ) : null}
          <button onClick={reset} style={{ marginTop: 16, padding: '8px 16px', cursor: 'pointer' }}>
            Try again
          </button>
        </div>
      </body>
    </html>
  );
}
