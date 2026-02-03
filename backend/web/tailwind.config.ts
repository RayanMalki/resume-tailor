import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: {
          950: "#0b0b0f",
          900: "#12121a",
          800: "#191925"
        },
        ember: {
          600: "#ff5b1a",
          500: "#ff7a1a",
          400: "#ff9850",
          300: "#ffb57f",
          200: "#ffd3ad"
        },
        glow: {
          orange: "#ff7a1a",
          red: "#ff3d3d",
          amber: "#ffb347"
        }
      },
      fontFamily: {
        grotesk: ["var(--font-grotesk)", "system-ui", "sans-serif"],
        plex: ["var(--font-plex)", "system-ui", "sans-serif"]
      },
      boxShadow: {
        card: "0 10px 30px rgba(15, 23, 42, 0.08)",
        glow: "0 12px 40px rgba(255, 122, 26, 0.35)",
        panel: "0 24px 60px rgba(0, 0, 0, 0.45)"
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
