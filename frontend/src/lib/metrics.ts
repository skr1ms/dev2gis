import { Counter, Histogram, register, collectDefaultMetrics } from 'prom-client';

let metricsInitialized = false;

export function initializeMetrics() {
  if (!metricsInitialized && typeof process !== 'undefined' && typeof process.uptime === 'function') {
    try {
      collectDefaultMetrics();
      metricsInitialized = true;
    } catch (error) {
      console.error('Failed to initialize metrics:', error);
    }
  }
}

export const httpRequestsTotal = new Counter({
  name: 'frontend_http_requests_total',
  help: 'Total number of HTTP requests',
  labelNames: ['method', 'path', 'status'],
});

export const httpRequestDuration = new Histogram({
  name: 'frontend_http_request_duration_seconds',
  help: 'Duration of HTTP requests in seconds',
  labelNames: ['method', 'path'],
  buckets: [0.1, 0.5, 1, 2, 5],
});

export const apiCallsTotal = new Counter({
  name: 'frontend_api_calls_total',
  help: 'Total number of API calls to backend',
  labelNames: ['endpoint', 'method', 'status'],
});

export const apiCallDuration = new Histogram({
  name: 'frontend_api_call_duration_seconds',
  help: 'Duration of API calls to backend',
  labelNames: ['endpoint', 'method'],
  buckets: [0.1, 0.5, 1, 2, 5, 10],
});

export { register };

