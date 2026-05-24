/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./src/**/*.{js,ts,jsx,tsx,mdx}'],
  theme: {
    extend: {
      colors: {
        ink: '#0f172a',
        muted: '#64748b',
        accent: '#2563eb',
        pass: '#16a34a',
        warn: '#ca8a04',
        fail: '#dc2626',
      },
    },
  },
  plugins: [],
};
