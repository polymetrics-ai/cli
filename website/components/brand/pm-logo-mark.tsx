type PmLogoMarkProps = {
  className?: string;
  decorative?: boolean;
  title?: string;
};

export function PmLogoMark({
  className = '',
  decorative = false,
  title = 'PM CLI logo',
}: PmLogoMarkProps) {
  const accessibilityProps = decorative
    ? { 'aria-hidden': true }
    : { role: 'img', 'aria-label': title };

  return (
    <svg
      className={className}
      viewBox="0 0 26 26"
      focusable="false"
      {...accessibilityProps}
    >
      <style>
        {`
          .pm-logo-mark__letter {
            fill: #fff;
            font: 700 13px var(--font-geist-mono), ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
            letter-spacing: -0.7px;
            text-rendering: geometricPrecision;
          }

          .pm-logo-mark__m {
            animation: pm-logo-mark-m 1.05s step-end infinite;
          }

          .pm-logo-mark__cursor {
            opacity: 0;
            animation: pm-logo-mark-cursor 1.05s step-end infinite;
          }

          @keyframes pm-logo-mark-m {
            0%, 49% { opacity: 1; }
            50%, 100% { opacity: 0; }
          }

          @keyframes pm-logo-mark-cursor {
            0%, 49% { opacity: 0; }
            50%, 100% { opacity: 1; }
          }

          @media (prefers-reduced-motion: reduce) {
            .pm-logo-mark__m,
            .pm-logo-mark__cursor {
              animation: none;
            }

            .pm-logo-mark__m {
              opacity: 1;
            }

            .pm-logo-mark__cursor {
              opacity: 0;
            }
          }
        `}
      </style>
      <rect width="26" height="26" fill="#065f46" />
      <text className="pm-logo-mark__letter" x="5.2" y="17.4">
        P
      </text>
      <text className="pm-logo-mark__letter pm-logo-mark__m" x="13.1" y="17.4">
        M
      </text>
      <text className="pm-logo-mark__letter pm-logo-mark__cursor" x="13.1" y="17.4">
        _
      </text>
    </svg>
  );
}
