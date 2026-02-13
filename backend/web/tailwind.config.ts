import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: {
          950: "#050505",
          900: "#0b0b0b",
          800: "#141414"
        },
        ember: {
          600: "#059669",
          500: "#10b981",
          400: "#34d399",
          300: "#6ee7b7",
          200: "#a7f3d0"
        },
        glow: {
          orange: "#10b981",
          red: "#34d399",
          amber: "#86efac"
        }
      },
      fontFamily: {
        grotesk: ["var(--font-grotesk)", "system-ui", "sans-serif"],
        plex: ["var(--font-plex)", "system-ui", "sans-serif"]
      },
      boxShadow: {
        card: "0 10px 30px rgba(15, 23, 42, 0.08)",
        glow: "0 10px 30px rgba(16, 185, 129, 0.22)",
        panel: "0 24px 60px rgba(0, 0, 0, 0.5)"
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
