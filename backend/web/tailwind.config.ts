import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: {
          950: "#0c1220",
          900: "#121a2b",
          800: "#1a2438"
        },
        ember: {
          600: "#1d4ed8",
          500: "#3b82f6",
          400: "#60a5fa",
          300: "#93c5fd",
          200: "#bfdbfe"
        },
        glow: {
          orange: "#3b82f6",
          red: "#2563eb",
          amber: "#22d3ee"
        }
      },
      fontFamily: {
        grotesk: ["var(--font-grotesk)", "system-ui", "sans-serif"],
        plex: ["var(--font-plex)", "system-ui", "sans-serif"]
      },
      boxShadow: {
        card: "0 10px 30px rgba(15, 23, 42, 0.08)",
        glow: "0 10px 30px rgba(59, 130, 246, 0.22)",
        panel: "0 18px 46px rgba(2, 6, 23, 0.35)"
      },
      keyframes: {
        caret: {
          "0%, 40%": { opacity: "1" },
          "41%, 100%": { opacity: "0" }
        }
      },
      animation: {
        caret: "caret 1.1s steps(1, end) infinite"
      },
      backgroundImage: {
        grid:
          "linear-gradient(transparent 0%, rgba(255,255,255,0.06) 1px, transparent 1px), linear-gradient(90deg, transparent 0%, rgba(255,255,255,0.06) 1px, transparent 1px)"
      }
    }
  },
  plugins: []
};

export default config;
