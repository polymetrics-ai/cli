import { ImageResponse } from 'next/og';

export const size = { width: 32, height: 32 };
export const contentType = 'image/png';

// Terminal-CLI favicon: white "PM" on an emerald-800 square.
export default function Icon() {
  return new ImageResponse(
    (
      <div
        style={{
          width: '100%',
          height: '100%',
          background: '#065f46', // emerald-800
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: '#ffffff',
          fontSize: 16,
          fontWeight: 700,
          fontFamily: 'monospace',
          letterSpacing: '-0.5px',
        }}
      >
        PM
      </div>
    ),
    { ...size },
  );
}
