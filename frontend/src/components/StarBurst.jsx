// An 8-point starburst used as the brand mark and as a faint card watermark.
// Color comes from CSS via currentColor.
export default function StarBurst({ className }) {
  return (
    <svg viewBox="0 0 100 100" className={className} aria-hidden="true" focusable="false">
      <path
        fill="currentColor"
        d="M50,2 L57.3,32.4 L83.9,16.1 L67.6,42.7 L98,50 L67.6,57.3 L83.9,83.9 L57.3,67.6 L50,98 L42.7,67.6 L16.1,83.9 L32.4,57.3 L2,50 L32.4,42.7 L16.1,16.1 L42.7,32.4 Z"
      />
    </svg>
  )
}
