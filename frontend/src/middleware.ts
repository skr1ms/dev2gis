import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export async function middleware(request: NextRequest) {
  const start = Date.now();
  const response = NextResponse.next();
  
  const duration = Date.now() - start;
  response.headers.set('X-Response-Time', `${duration}ms`);
  response.headers.set('X-Request-Path', request.nextUrl.pathname);
  
  return response;
}

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|api/track|api/metrics).*)',
  ],
};

