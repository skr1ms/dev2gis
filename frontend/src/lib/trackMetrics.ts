export async function trackPageView(path: string, method: string = 'GET') {
  const start = Date.now();
  
  return () => {
    const duration = Date.now() - start;
    
    if (typeof window !== 'undefined') {
      fetch('/api/track', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          method,
          path,
          status: 200,
          duration,
        }),
      }).catch(err => console.error('Failed to track metrics:', err));
    }
  };
}

