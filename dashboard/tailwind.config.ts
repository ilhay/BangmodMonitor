import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        bg: {
          DEFAULT: '#050d1a',
          card: '#0c1831',
          hover: '#0f1e38',
          sidebar: '#08132b',
          input: '#060e1d',
        },
        border: {
          DEFAULT: '#1a2d4a',
          muted: '#0f1e35',
          focus: '#3b82f6',
        },
        accent: {
          DEFAULT: '#3b82f6',
          hover: '#60a5fa',
          muted: 'rgba(59,130,246,0.12)',
          glow: 'rgba(59,130,246,0.3)',
        },
        success: { DEFAULT: '#22c55e', muted: 'rgba(34,197,94,0.1)' },
        warning: { DEFAULT: '#f59e0b', muted: 'rgba(245,158,11,0.1)' },
        danger: { DEFAULT: '#ef4444', muted: 'rgba(239,68,68,0.1)' },
      },
      fontFamily: {
        sans: ['var(--font-geist-sans)'],
        mono: ['var(--font-geist-mono)'],
      },
      boxShadow: {
        card: '0 1px 3px rgba(0,0,0,0.4), 0 0 0 1px rgba(26,45,74,0.8)',
        glow: '0 0 20px rgba(59,130,246,0.2)',
        'glow-sm': '0 0 8px rgba(59,130,246,0.3)',
      },
      animation: {
        'pulse-dot': 'pulse-dot 2s cubic-bezier(0.4,0,0.6,1) infinite',
        'fade-in': 'fade-in 0.3s ease',
      },
      keyframes: {
        'pulse-dot': {
          '0%,100%': { opacity: '1' },
          '50%': { opacity: '0.4' },
        },
        'fade-in': {
          from: { opacity: '0', transform: 'translateY(4px)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
      },
    },
  },
  plugins: [],
}

export default config
