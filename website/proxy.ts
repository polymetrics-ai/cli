import { NextRequest, NextResponse } from 'next/server';

const HSTS_HEADER = 'max-age=63072000; includeSubDomains; preload';

function forwardedProto(request: NextRequest) {
  return request.headers.get('x-forwarded-proto')?.split(',')[0]?.trim().toLowerCase() ?? '';
}

function hostnameFromHost(host: string) {
  if (host.startsWith('[')) {
    return host.slice(1, host.indexOf(']')).toLowerCase();
  }

  return host.split(':')[0]?.toLowerCase() ?? '';
}

function isLocalHost(host: string) {
  const hostname = hostnameFromHost(host);
  return hostname === 'localhost' || hostname === '127.0.0.1' || hostname === '::1';
}

export function proxy(request: NextRequest) {
  const proto = forwardedProto(request);
  const host = request.headers.get('host') ?? request.nextUrl.host;

  if (proto === 'http' && !isLocalHost(host)) {
    const url = request.nextUrl.clone();
    url.protocol = 'https:';
    url.hostname = hostnameFromHost(host);
    url.port = '';
    return NextResponse.redirect(url, 308);
  }

  const response = NextResponse.next();

  if (proto === 'https' || request.nextUrl.protocol === 'https:') {
    response.headers.set('strict-transport-security', HSTS_HEADER);
  }

  return response;
}

export const config = {
  matcher: '/:path*',
};
