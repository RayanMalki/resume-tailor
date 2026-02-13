"use client";

import { Component, ReactNode } from "react";

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
}

/**
 * Catches unhandled React rendering errors and shows a fallback UI
 * instead of crashing the entire page with a white screen.
 */
export default class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true };
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    // You can send this to an error tracking service (e.g. Sentry) later.
    console.error("ErrorBoundary caught:", error, info.componentStack);
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="flex min-h-[60vh] items-center justify-center px-6">
          <div className="max-w-md rounded-2xl border border-white/10 bg-ink-900/70 p-8 text-center shadow-panel backdrop-blur">
            <h2 className="text-xl font-semibold text-white">Something went wrong</h2>
            <p className="mt-3 text-sm text-slate-400">
              An unexpected error occurred. Please try refreshing the page.
            </p>
            <button
              onClick={() => {
                this.setState({ hasError: false });
                window.location.reload();
              }}
              className="mt-6 rounded-full border border-ember-500/60 bg-ink-900/60 px-6 py-2 text-sm font-semibold uppercase tracking-[0.15em] text-ember-200 transition hover:border-ember-400 hover:text-white"
            >
              Refresh page
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
