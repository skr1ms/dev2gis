import { NextRequest, NextResponse } from 'next/server';
import { register, initializeMetrics } from '@/lib/metrics';

export const runtime = 'nodejs';

export async function GET(request: NextRequest) {
  try {
    initializeMetrics();
    
    const metrics = await register.metrics();
    return new NextResponse(metrics, {
      headers: {
        'Content-Type': register.contentType,
      },
    });
  } catch (error) {
    console.error('Error generating metrics:', error);
    return new NextResponse('Error generating metrics', { status: 500 });
  }
}

