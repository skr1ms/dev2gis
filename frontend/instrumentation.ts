export async function register() {
  if (process.env.NEXT_RUNTIME === 'nodejs') {
    const { initializeMetrics } = await import('./src/lib/metrics');
    initializeMetrics();
    console.log('Metrics initialized in Node.js runtime');
  }
}

