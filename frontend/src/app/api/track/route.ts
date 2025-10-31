import { NextRequest, NextResponse } from 'next/server';
import { httpRequestsTotal, httpRequestDuration } from '@/lib/metrics';
import logger from '@/lib/logger';

export const runtime = 'nodejs';

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { method, path, status, duration } = body;

    if (method && path && status && duration !== undefined) {
      httpRequestsTotal.labels(method, path, status.toString()).inc();
      httpRequestDuration.labels(method, path).observe(duration / 1000);
      
      logger.info('HTTP Request', {
        method,
        path,
        status,
        latency: `${duration}ms`,
        user_agent: request.headers.get('user-agent'),
      });
    }

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error('Error tracking request:', error);
    return NextResponse.json({ success: false }, { status: 500 });
  }
}

